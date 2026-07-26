package postgres

import (
	"context"
	"database/sql"
	"fmt"

	"awesomeProject/internal/repository"
)

type CartRepository struct {
	db *sql.DB
}

func NewCartRepository(db *sql.DB) *CartRepository {
	return &CartRepository{db: db}
}

func (r *CartRepository) ensureCartRow(ctx context.Context, tx *sql.Tx, userID string) error {
	_, err := tx.ExecContext(ctx, `
		INSERT INTO carts (user_id, promocode, updated_at)
		VALUES ($1, NULL, NOW())
		ON CONFLICT (user_id) DO NOTHING
	`, userID)
	return err
}

func (r *CartRepository) GetCart(userID string) (repository.Cart, error) {
	ctx := context.Background()

	cart := repository.Cart{
		UserID: userID,
		Items:  make(map[string]repository.CartItem),
	}

	const metaQ = `
		SELECT promocode, updated_at
		FROM carts
		WHERE user_id = $1
	`
	var promo sql.NullString
	var updatedAt sql.NullTime
	err := r.db.QueryRowContext(ctx, metaQ, userID).Scan(&promo, &updatedAt)
	if err != nil && err != sql.ErrNoRows {
		return repository.Cart{}, fmt.Errorf("get cart meta: %w", err)
	}
	if promo.Valid {
		cart.Promocode = promo.String
	}
	if updatedAt.Valid {
		cart.UpdatedAt = updatedAt.Time
	}

	const q = `
		SELECT product_id, quantity
		FROM cart_items
		WHERE user_id = $1
	`
	rows, err := r.db.QueryContext(ctx, q, userID)
	if err != nil {
		return repository.Cart{}, fmt.Errorf("get cart: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var item repository.CartItem
		if err := rows.Scan(&item.ProductID, &item.Quantity); err != nil {
			return repository.Cart{}, fmt.Errorf("scan cart item: %w", err)
		}
		cart.Items[item.ProductID] = item
	}
	if err := rows.Err(); err != nil {
		return repository.Cart{}, fmt.Errorf("get cart rows: %w", err)
	}

	return cart, nil
}

func (r *CartRepository) SaveCart(cart repository.Cart) error {
	ctx := context.Background()
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin save cart: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	if err := r.ensureCartRow(ctx, tx, cart.UserID); err != nil {
		return fmt.Errorf("ensure cart row: %w", err)
	}

	if _, err := tx.ExecContext(ctx, `DELETE FROM cart_items WHERE user_id = $1`, cart.UserID); err != nil {
		return fmt.Errorf("clear cart before save: %w", err)
	}

	const insertQ = `
		INSERT INTO cart_items (user_id, product_id, quantity)
		VALUES ($1, $2, $3)
	`
	for _, item := range cart.Items {
		if _, err := tx.ExecContext(ctx, insertQ, cart.UserID, item.ProductID, item.Quantity); err != nil {
			return fmt.Errorf("insert cart item: %w", err)
		}
	}

	if _, err := tx.ExecContext(ctx, `
		UPDATE carts SET updated_at = NOW() WHERE user_id = $1
	`, cart.UserID); err != nil {
		return fmt.Errorf("touch cart: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit save cart: %w", err)
	}
	return nil
}

func (r *CartRepository) ClearCart(userID string) error {
	ctx := context.Background()
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin clear cart: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	if _, err := tx.ExecContext(ctx, `DELETE FROM cart_items WHERE user_id = $1`, userID); err != nil {
		return fmt.Errorf("clear cart: %w", err)
	}

	if err := r.ensureCartRow(ctx, tx, userID); err != nil {
		return fmt.Errorf("ensure cart row: %w", err)
	}

	if _, err := tx.ExecContext(ctx, `
		UPDATE carts SET updated_at = NOW() WHERE user_id = $1
	`, userID); err != nil {
		return fmt.Errorf("touch cart on clear: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit clear cart: %w", err)
	}
	return nil
}

func (r *CartRepository) SetPromocode(userID, code string) error {
	ctx := context.Background()
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO carts (user_id, promocode, updated_at)
		VALUES ($1, $2, NOW())
		ON CONFLICT (user_id) DO UPDATE
		SET promocode = EXCLUDED.promocode, updated_at = NOW()
	`, userID, code)
	if err != nil {
		return fmt.Errorf("set promocode: %w", err)
	}
	return nil
}

func (r *CartRepository) ClearPromocode(userID string) error {
	ctx := context.Background()
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO carts (user_id, promocode, updated_at)
		VALUES ($1, NULL, NOW())
		ON CONFLICT (user_id) DO UPDATE
		SET promocode = NULL, updated_at = NOW()
	`, userID)
	if err != nil {
		return fmt.Errorf("clear promocode: %w", err)
	}
	return nil
}

func (r *CartRepository) TouchCart(userID string) error {
	ctx := context.Background()
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO carts (user_id, promocode, updated_at)
		VALUES ($1, NULL, NOW())
		ON CONFLICT (user_id) DO UPDATE
		SET updated_at = NOW()
	`, userID)
	if err != nil {
		return fmt.Errorf("touch cart: %w", err)
	}
	return nil
}
