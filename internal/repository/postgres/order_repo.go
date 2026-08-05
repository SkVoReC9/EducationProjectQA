package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"

	"awesomeProject/internal/repository"
)

type OrderRepository struct {
	db *sql.DB
}

func NewOrderRepository(db *sql.DB) *OrderRepository {
	return &OrderRepository{db: db}
}

func (r *OrderRepository) CreateOrder(order repository.Order) (repository.Order, error) {
	ctx := context.Background()
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return repository.Order{}, fmt.Errorf("begin create order: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	order.ID = uuid.NewString()
	order.Status = 1 // ORDER_STATUS_CREATED
	now := time.Now()
	order.CreatedAt = now
	order.UpdatedAt = now

	const insertOrder = `
		INSERT INTO orders (id, user_id, total_amount_cents, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`
	if _, err := tx.ExecContext(ctx, insertOrder, order.ID, order.UserID, order.TotalAmountCents, order.Status, order.CreatedAt, order.UpdatedAt); err != nil {
		return repository.Order{}, fmt.Errorf("insert order: %w", err)
	}

	const insertItem = `
		INSERT INTO order_items (order_id, product_id, quantity, price_cents)
		VALUES ($1, $2, $3, $4)
	`
	for _, item := range order.Items {
		if _, err := tx.ExecContext(ctx, insertItem, order.ID, item.ProductID, item.Quantity, item.PriceCents); err != nil {
			return repository.Order{}, fmt.Errorf("insert order item: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return repository.Order{}, fmt.Errorf("commit create order: %w", err)
	}
	return order, nil
}

func (r *OrderRepository) GetOrder(orderID string) (repository.Order, error) {
	ctx := context.Background()

	const orderQ = `
		SELECT id, user_id, total_amount_cents, status, created_at, updated_at
		FROM orders
		WHERE id = $1
	`

	var order repository.Order
	err := r.db.QueryRowContext(ctx, orderQ, orderID).Scan(
		&order.ID,
		&order.UserID,
		&order.TotalAmountCents,
		&order.Status,
		&order.CreatedAt,
		&order.UpdatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return repository.Order{}, repository.ErrOrderNotFound
	}
	if err != nil {
		return repository.Order{}, fmt.Errorf("get order: %w", err)
	}

	items, err := r.loadItems(ctx, orderID)
	if err != nil {
		return repository.Order{}, err
	}
	order.Items = items
	return order, nil
}

func (r *OrderRepository) loadItems(ctx context.Context, orderID string) ([]repository.OrderItem, error) {
	const itemsQ = `
		SELECT product_id, quantity, price_cents
		FROM order_items
		WHERE order_id = $1
		ORDER BY product_id
	`
	rows, err := r.db.QueryContext(ctx, itemsQ, orderID)
	if err != nil {
		return nil, fmt.Errorf("get order items: %w", err)
	}
	defer rows.Close()

	items := make([]repository.OrderItem, 0)
	for rows.Next() {
		var item repository.OrderItem
		if err := rows.Scan(&item.ProductID, &item.Quantity, &item.PriceCents); err != nil {
			return nil, fmt.Errorf("scan order item: %w", err)
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("get order items rows: %w", err)
	}
	return items, nil
}

func (r *OrderRepository) UpdateOrderStatus(orderID string, fromStatus, toStatus int32) (repository.Order, error) {
	ctx := context.Background()
	const q = `
		UPDATE orders
		SET status = $3, updated_at = NOW()
		WHERE id = $1 AND status = $2
		RETURNING id, user_id, total_amount_cents, status, created_at, updated_at
	`
	var order repository.Order
	err := r.db.QueryRowContext(ctx, q, orderID, fromStatus, toStatus).Scan(
		&order.ID,
		&order.UserID,
		&order.TotalAmountCents,
		&order.Status,
		&order.CreatedAt,
		&order.UpdatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return repository.Order{}, fmt.Errorf("status transition rejected")
	}
	if err != nil {
		return repository.Order{}, fmt.Errorf("update order status: %w", err)
	}
	items, err := r.loadItems(ctx, orderID)
	if err != nil {
		return repository.Order{}, err
	}
	order.Items = items
	return order, nil
}

func (r *OrderRepository) DeleteOrder(orderID string) error {
	res, err := r.db.ExecContext(context.Background(), `DELETE FROM orders WHERE id = $1`, orderID)
	if err != nil {
		return fmt.Errorf("delete order: %w", err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return repository.ErrOrderNotFound
	}
	return nil
}
