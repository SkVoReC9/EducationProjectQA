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

func (r *CartRepository) GetCart(userID string) (repository.Cart, error) {
	const q = `
		SELECT product_id, quantity
		FROM cart_items
		WHERE user_id = $1
	`

	rows, err := r.db.QueryContext(context.Background(), q, userID)
	if err != nil {
		return repository.Cart{}, fmt.Errorf("get cart: %w", err)
	}
	defer rows.Close()

	cart := repository.Cart{
		UserID: userID,
		Items:  make(map[string]repository.CartItem),
	}

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

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit save cart: %w", err)
	}
	return nil
}

func (r *CartRepository) ClearCart(userID string) error {
	_, err := r.db.ExecContext(context.Background(), `DELETE FROM cart_items WHERE user_id = $1`, userID)
	if err != nil {
		return fmt.Errorf("clear cart: %w", err)
	}
	return nil
}
