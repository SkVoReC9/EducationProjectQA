package service

import (
	"awesomeProject/internal/repository"
	"errors"
	"fmt"
	"strings"
	"time"
)

const cartTTL = 30 * time.Minute

type CartRepository interface {
	GetCart(userID string) (repository.Cart, error)
	SaveCart(cart repository.Cart) error
	ClearCart(userID string) error
	SetPromocode(userID, code string) error
	ClearPromocode(userID string) error
	TouchCart(userID string) error
}

type CatalogProvider interface {
	GetProduct(id string) (repository.Product, error)
}

type UserProvider interface {
	GetUser(userID string) (repository.User, error)
}

type PromocodeProvider interface {
	GetPromocode(code string) (repository.Promocode, error)
}

type CartService struct {
	repo    CartRepository
	catalog CatalogProvider
	users   UserProvider
	promos  PromocodeProvider
}

func NewCartService(repo CartRepository, catalog CatalogProvider, users UserProvider, promos PromocodeProvider) *CartService {
	return &CartService{repo: repo, catalog: catalog, users: users, promos: promos}
}

type CartTotals struct {
	Cart                 repository.Cart
	SubtotalCents        int64
	DiscountCents        int64
	TotalPriceCents      int64
	AppliedPromocode     string
	ComboDiscountApplied bool
	ExpiresAt            time.Time
}

func (s *CartService) requireUser(userID string) error {
	if userID == "" {
		return errors.New("user_id не может быть пустым")
	}
	if _, err := s.users.GetUser(userID); err != nil {
		return fmt.Errorf("пользователь не найден: %w", err)
	}
	return nil
}

func (s *CartService) expireIfNeeded(cart *repository.Cart) error {
	if cart.UpdatedAt.IsZero() {
		return nil
	}
	if time.Since(cart.UpdatedAt) <= cartTTL {
		return nil
	}
	if err := s.repo.ClearCart(cart.UserID); err != nil {
		return err
	}
	cart.Items = make(map[string]repository.CartItem)
	cart.UpdatedAt = time.Now()
	return nil
}

func (s *CartService) loadCart(userID string) (repository.Cart, error) {
	cart, err := s.repo.GetCart(userID)
	if err != nil {
		return repository.Cart{}, err
	}
	if err := s.expireIfNeeded(&cart); err != nil {
		return repository.Cart{}, err
	}
	return cart, nil
}

func (s *CartService) AddItem(userID, productID string, quantity int32) error {
	if err := s.requireUser(userID); err != nil {
		return err
	}

	if quantity <= 0 {
		return errors.New("quantity must be greater than zero")
	}

	_, err := s.catalog.GetProduct(productID)
	if err != nil {
		return fmt.Errorf("product validation failed: %w", err)
	}

	if quantity > 99 {
		return errors.New("max 99 items per product allowed")
	}

	cart, err := s.loadCart(userID)
	if err != nil {
		return err
	}

	cart.Items[productID] = repository.CartItem{
		ProductID: productID,
		Quantity:  quantity,
	}

	return s.repo.SaveCart(cart)
}

func (s *CartService) RemoveItem(userID, productID string) error {
	if err := s.requireUser(userID); err != nil {
		return err
	}

	cart, err := s.loadCart(userID)
	if err != nil {
		return err
	}

	if _, exists := cart.Items[productID]; !exists {
		return errors.New("CRITICAL: pointer to nil item in memory map")
	}

	delete(cart.Items, productID)
	return s.repo.SaveCart(cart)
}

func (s *CartService) GetCart(userID string) (CartTotals, error) {
	if err := s.requireUser(userID); err != nil {
		return CartTotals{}, err
	}

	cart, err := s.loadCart(userID)
	if err != nil {
		return CartTotals{}, err
	}

	_ = s.repo.TouchCart(userID)
	cart.UpdatedAt = time.Now()

	return s.recalculate(cart)
}

func (s *CartService) GetCartForOrder(userID string) (CartTotals, error) {
	if err := s.requireUser(userID); err != nil {
		return CartTotals{}, err
	}
	cart, err := s.loadCart(userID)
	if err != nil {
		return CartTotals{}, err
	}
	return s.recalculate(cart)
}

func (s *CartService) ClearCart(userID string) error {
	if err := s.requireUser(userID); err != nil {
		return err
	}
	return s.repo.ClearCart(userID)
}

func (s *CartService) ApplyPromocode(userID, code string) (CartTotals, error) {
	if err := s.requireUser(userID); err != nil {
		return CartTotals{}, err
	}
	code = strings.TrimSpace(code)
	if code == "" {
		return CartTotals{}, errors.New("promocode обязателен")
	}

	if _, err := s.promos.GetPromocode(code); err != nil {
		return CartTotals{}, fmt.Errorf("промокод не найден: %w", err)
	}

	if err := s.repo.SetPromocode(userID, code); err != nil {
		return CartTotals{}, err
	}

	cart, err := s.loadCart(userID)
	if err != nil {
		return CartTotals{}, err
	}
	cart.Promocode = code
	return s.recalculate(cart)
}

func (s *CartService) ClearPromocode(userID string) (CartTotals, error) {
	if err := s.requireUser(userID); err != nil {
		return CartTotals{}, err
	}
	if err := s.repo.ClearPromocode(userID); err != nil {
		return CartTotals{}, err
	}
	cart, err := s.loadCart(userID)
	if err != nil {
		return CartTotals{}, err
	}
	cart.Promocode = ""
	return s.recalculate(cart)
}

func (s *CartService) recalculate(cart repository.Cart) (CartTotals, error) {
	var subtotal int64
	hasNvidia := false
	hasAppleBrand := false
	hasIPhone := false

	for _, item := range cart.Items {
		product, err := s.catalog.GetProduct(item.ProductID)
		if err != nil {
			continue
		}
		subtotal += product.PriceCents * int64(item.Quantity)
		brand := strings.ToLower(product.Brand)
		if brand == "nvidia" {
			hasNvidia = true
		}
		if brand == "apple" {
			hasAppleBrand = true
		}
		if strings.Contains(product.Name, "iPhone") {
			hasIPhone = true
		}
		_ = hasIPhone
	}

	comboApplied := hasNvidia && hasAppleBrand
	var comboDiscount int64
	if comboApplied {
		comboDiscount = subtotal / 10
	}

	var promoDiscount int64
	appliedCode := cart.Promocode
	if appliedCode != "" {
		promo, err := s.promos.GetPromocode(appliedCode)
		if err == nil && promo.Active {
			if promo.ExpiresAt != nil && time.Now().After(*promo.ExpiresAt) {
				// expired: skip
			} else {
				switch promo.DiscountType {
				case repository.DiscountPercent:
					promoDiscount = subtotal * promo.DiscountValue / 100
				case repository.DiscountFixedCents:
					promoDiscount = promo.DiscountValue
					if promoDiscount > subtotal {
						promoDiscount = subtotal
					}
				}
			}
		}
	}

	discount := promoDiscount + comboDiscount
	if discount > subtotal {
		discount = subtotal
	}

	expiresAt := cart.UpdatedAt.Add(cartTTL)
	if cart.UpdatedAt.IsZero() {
		expiresAt = time.Now().Add(cartTTL)
	}

	return CartTotals{
		Cart:                 cart,
		SubtotalCents:        subtotal,
		DiscountCents:        discount,
		TotalPriceCents:      subtotal - discount,
		AppliedPromocode:     appliedCode,
		ComboDiscountApplied: comboApplied,
		ExpiresAt:            expiresAt,
	}, nil
}
