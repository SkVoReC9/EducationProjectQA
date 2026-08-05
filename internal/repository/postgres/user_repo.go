package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"

	"awesomeProject/internal/repository"
)

type UserRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) CreateUser(user repository.User) (repository.User, error) {
	if user.ID == "" {
		user.ID = uuid.NewString()
	}
	if user.Role == "" {
		user.Role = "user"
	}

	const q = `
		INSERT INTO users (id, email, password_hash, name, role)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, email, password_hash, name, role, created_at
	`

	var saved repository.User
	err := r.db.QueryRowContext(
		context.Background(),
		q,
		user.ID,
		user.Email,
		user.PasswordHash,
		user.Name,
		user.Role,
	).Scan(&saved.ID, &saved.Email, &saved.PasswordHash, &saved.Name, &saved.Role, &saved.CreatedAt)
	if err != nil {
		return repository.User{}, fmt.Errorf("create user: %w", err)
	}
	return saved, nil
}

func (r *UserRepository) GetUser(id string) (repository.User, error) {
	const q = `
		SELECT id, email, password_hash, name, role, created_at
		FROM users
		WHERE id = $1
	`

	var user repository.User
	err := r.db.QueryRowContext(context.Background(), q, id).
		Scan(&user.ID, &user.Email, &user.PasswordHash, &user.Name, &user.Role, &user.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return repository.User{}, repository.ErrUserNotFound
	}
	if err != nil {
		return repository.User{}, fmt.Errorf("get user: %w", err)
	}
	return user, nil
}

func (r *UserRepository) GetUserByEmail(email string) (repository.User, error) {
	const q = `
		SELECT id, email, password_hash, name, role, created_at
		FROM users
		WHERE email = $1
	`

	var user repository.User
	err := r.db.QueryRowContext(context.Background(), q, email).
		Scan(&user.ID, &user.Email, &user.PasswordHash, &user.Name, &user.Role, &user.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return repository.User{}, repository.ErrUserNotFound
	}
	if err != nil {
		return repository.User{}, fmt.Errorf("get user by email: %w", err)
	}
	return user, nil
}

func (r *UserRepository) DeleteUser(id string) error {
	ctx := context.Background()
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin delete user: %w", err)
	}
	defer tx.Rollback()

	var activeCount int
	err = tx.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM orders
		WHERE user_id = $1
		  AND status NOT IN ($2, $3)
	`, id, repository.OrderStatusCancelled, repository.OrderStatusCompleted).Scan(&activeCount)
	if err != nil {
		return fmt.Errorf("count active orders: %w", err)
	}
	if activeCount > 0 {
		return repository.ErrUserHasActiveOrders
	}

	// order_items cascade on order delete
	if _, err := tx.ExecContext(ctx, `
		DELETE FROM orders
		WHERE user_id = $1
		  AND status IN ($2, $3)
	`, id, repository.OrderStatusCancelled, repository.OrderStatusCompleted); err != nil {
		return fmt.Errorf("delete terminal orders: %w", err)
	}

	res, err := tx.ExecContext(ctx, `DELETE FROM users WHERE id = $1`, id)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23503" {
			return repository.ErrUserHasActiveOrders
		}
		return fmt.Errorf("delete user: %w", err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return repository.ErrUserNotFound
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit delete user: %w", err)
	}
	return nil
}
