package postgres

import (
	"context"
	"database/sql"
	"fmt"

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

	const insertOrder = `
		INSERT INTO orders (id, user_id, total_amount_cents, status)
		VALUES ($1, $2, $3, $4)
	`
	if _, err := tx.ExecContext(ctx, insertOrder, order.ID, order.UserID, order.TotalAmountCents, order.Status); err != nil {
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
		SELECT id, user_id, total_amount_cents, status
		FROM orders
		WHERE id = $1
	`

	var order repository.Order
	err := r.db.QueryRowContext(ctx, orderQ, orderID).Scan(
		&order.ID,
		&order.UserID,
		&order.TotalAmountCents,
		&order.Status,
	)
	if err == sql.ErrNoRows {
		return repository.Order{}, fmt.Errorf("заказ не найден")
	}
	if err != nil {
		return repository.Order{}, fmt.Errorf("get order: %w", err)
	}

	const itemsQ = `
		SELECT product_id, quantity, price_cents
		FROM order_items
		WHERE order_id = $1
		ORDER BY product_id
	`
	rows, err := r.db.QueryContext(ctx, itemsQ, orderID)
	if err != nil {
		return repository.Order{}, fmt.Errorf("get order items: %w", err)
	}
	defer rows.Close()

	order.Items = make([]repository.OrderItem, 0)
	for rows.Next() {
		var item repository.OrderItem
		if err := rows.Scan(&item.ProductID, &item.Quantity, &item.PriceCents); err != nil {
			return repository.Order{}, fmt.Errorf("scan order item: %w", err)
		}
		order.Items = append(order.Items, item)
	}
	if err := rows.Err(); err != nil {
		return repository.Order{}, fmt.Errorf("get order items rows: %w", err)
	}

	return order, nil
}
