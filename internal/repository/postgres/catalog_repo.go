package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"awesomeProject/internal/repository"
)

type CatalogRepository struct {
	db *sql.DB
}

func NewCatalogRepository(db *sql.DB) *CatalogRepository {
	return &CatalogRepository{db: db}
}

func (r *CatalogRepository) GetProduct(id string) (repository.Product, error) {
	const q = `
		SELECT id, name, description, price_cents, stock_quantity
		FROM products
		WHERE id = $1
	`

	var p repository.Product
	err := r.db.QueryRowContext(context.Background(), q, id).Scan(
		&p.ID,
		&p.Name,
		&p.Description,
		&p.PriceCents,
		&p.StockQuantity,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return repository.Product{}, repository.ErrNotFound
	}
	if err != nil {
		return repository.Product{}, fmt.Errorf("get product: %w", err)
	}
	return p, nil
}

func (r *CatalogRepository) ListProducts() ([]repository.Product, error) {
	const q = `
		SELECT id, name, description, price_cents, stock_quantity
		FROM products
		ORDER BY id
	`

	rows, err := r.db.QueryContext(context.Background(), q)
	if err != nil {
		return nil, fmt.Errorf("list products: %w", err)
	}
	defer rows.Close()

	products := make([]repository.Product, 0)
	for rows.Next() {
		var p repository.Product
		if err := rows.Scan(&p.ID, &p.Name, &p.Description, &p.PriceCents, &p.StockQuantity); err != nil {
			return nil, fmt.Errorf("scan product: %w", err)
		}
		products = append(products, p)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list products rows: %w", err)
	}
	return products, nil
}
