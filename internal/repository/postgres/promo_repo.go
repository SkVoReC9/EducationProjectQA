package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"awesomeProject/internal/repository"
)

type PromoRepository struct {
	db *sql.DB
}

func NewPromoRepository(db *sql.DB) *PromoRepository {
	return &PromoRepository{db: db}
}

func scanPromocode(scanner interface {
	Scan(dest ...any) error
}) (repository.Promocode, error) {
	var p repository.Promocode
	var expiresAt sql.NullTime
	if err := scanner.Scan(&p.Code, &p.DiscountType, &p.DiscountValue, &p.Active, &expiresAt); err != nil {
		return repository.Promocode{}, err
	}
	if expiresAt.Valid {
		t := expiresAt.Time
		p.ExpiresAt = &t
	}
	return p, nil
}

func (r *PromoRepository) GetPromocode(code string) (repository.Promocode, error) {
	const q = `
		SELECT code, discount_type, discount_value, active, expires_at
		FROM promocodes
		WHERE code = $1
	`
	p, err := scanPromocode(r.db.QueryRowContext(context.Background(), q, code))
	if errors.Is(err, sql.ErrNoRows) {
		return repository.Promocode{}, repository.ErrPromocodeNotFound
	}
	if err != nil {
		return repository.Promocode{}, fmt.Errorf("get promocode: %w", err)
	}
	return p, nil
}

func (r *PromoRepository) ListPromocodes() ([]repository.Promocode, error) {
	const q = `
		SELECT code, discount_type, discount_value, active, expires_at
		FROM promocodes
		ORDER BY code
	`
	rows, err := r.db.QueryContext(context.Background(), q)
	if err != nil {
		return nil, fmt.Errorf("list promocodes: %w", err)
	}
	defer rows.Close()

	out := make([]repository.Promocode, 0)
	for rows.Next() {
		p, err := scanPromocode(rows)
		if err != nil {
			return nil, fmt.Errorf("scan promocode: %w", err)
		}
		out = append(out, p)
	}
	return out, rows.Err()
}

func (r *PromoRepository) CreatePromocode(p repository.Promocode) (repository.Promocode, error) {
	const q = `
		INSERT INTO promocodes (code, discount_type, discount_value, active, expires_at)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING code, discount_type, discount_value, active, expires_at
	`
	var expires any
	if p.ExpiresAt != nil {
		expires = *p.ExpiresAt
	}
	saved, err := scanPromocode(r.db.QueryRowContext(
		context.Background(), q, p.Code, p.DiscountType, p.DiscountValue, p.Active, expires,
	))
	if err != nil {
		return repository.Promocode{}, fmt.Errorf("create promocode: %w", err)
	}
	return saved, nil
}

func (r *PromoRepository) UpdatePromocode(p repository.Promocode) (repository.Promocode, error) {
	const q = `
		UPDATE promocodes
		SET discount_type = $2, discount_value = $3, active = $4, expires_at = $5
		WHERE code = $1
		RETURNING code, discount_type, discount_value, active, expires_at
	`
	var expires any
	if p.ExpiresAt != nil {
		expires = *p.ExpiresAt
	}
	saved, err := scanPromocode(r.db.QueryRowContext(
		context.Background(), q, p.Code, p.DiscountType, p.DiscountValue, p.Active, expires,
	))
	if errors.Is(err, sql.ErrNoRows) {
		return repository.Promocode{}, repository.ErrPromocodeNotFound
	}
	if err != nil {
		return repository.Promocode{}, fmt.Errorf("update promocode: %w", err)
	}
	return saved, nil
}

func (r *PromoRepository) DeletePromocode(code string) error {
	res, err := r.db.ExecContext(context.Background(), `DELETE FROM promocodes WHERE code = $1`, code)
	if err != nil {
		return fmt.Errorf("delete promocode: %w", err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return repository.ErrPromocodeNotFound
	}
	return nil
}
