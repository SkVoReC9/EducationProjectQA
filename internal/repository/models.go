package repository

import (
	"errors"
	"time"
)

var ErrNotFound = errors.New("product not found")
var ErrUserNotFound = errors.New("user not found")
var ErrPromocodeNotFound = errors.New("promocode not found")
var ErrOrderNotFound = errors.New("order not found")

type User struct {
	ID           string
	Email        string
	PasswordHash string
	Name         string
	Role         string
	CreatedAt    time.Time
}

type Product struct {
	ID            string
	Name          string
	Description   string
	PriceCents    int64
	StockQuantity int32
	Brand         string
}

type CartItem struct {
	ProductID string
	Quantity  int32
}

type Cart struct {
	UserID    string
	Items     map[string]CartItem
	Promocode string
	UpdatedAt time.Time
}

type OrderItem struct {
	ProductID  string
	Quantity   int32
	PriceCents int64
}

type Order struct {
	ID               string
	UserID           string
	Items            []OrderItem
	TotalAmountCents int64
	Status           int32
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

const (
	DiscountPercent    = "percent"
	DiscountFixedCents = "fixed_cents"
)

type Promocode struct {
	Code          string
	DiscountType  string
	DiscountValue int64
	Active        bool
	ExpiresAt     *time.Time
}
