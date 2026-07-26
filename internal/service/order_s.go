package service

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"awesomeProject/internal/repository"
)

const orderStatusTTL = 10 * time.Minute

const (
	OrderStatusCreated   int32 = 1
	OrderStatusPaid      int32 = 2
	OrderStatusShipped   int32 = 3
	OrderStatusCancelled int32 = 4
	OrderStatusCompleted int32 = 5
)

var (
	ErrOrderNotFound       = repository.ErrOrderNotFound
	ErrInvalidTransition   = errors.New("недопустимый переход статуса")
	ErrStatusMismatch      = errors.New("текущий статус не совпадает с from_status")
	ErrPermissionDenied    = errors.New("нет доступа к заказу")
)

type OrderRepository interface {
	CreateOrder(order repository.Order) (repository.Order, error)
	GetOrder(orderID string) (repository.Order, error)
	UpdateOrderStatus(orderID string, fromStatus, toStatus int32) (repository.Order, error)
}

type CartProvider interface {
	GetCartForOrder(userID string) (CartTotals, error)
	ClearCart(userID string) error
}

type OrderService struct {
	repo    OrderRepository
	cart    CartProvider
	catalog CatalogProvider
	users   UserProvider
}

func NewOrderService(repo OrderRepository, cart CartProvider, catalog CatalogProvider, users UserProvider) *OrderService {
	return &OrderService{repo: repo, cart: cart, catalog: catalog, users: users}
}

func (s *OrderService) CreateOrder(userID string) (repository.Order, error) {
	if userID == "" {
		return repository.Order{}, errors.New("user_id не может быть пустым")
	}
	if _, err := s.users.GetUser(userID); err != nil {
		return repository.Order{}, fmt.Errorf("пользователь не найден: %w", err)
	}

	totals, err := s.cart.GetCartForOrder(userID)
	if err != nil {
		return repository.Order{}, fmt.Errorf("не удалось получить корзину: %v", err)
	}

	var orderItems []repository.OrderItem
	var actualTotal int64

	for _, cartItem := range totals.Cart.Items {
		product, err := s.catalog.GetProduct(cartItem.ProductID)
		if err != nil {
			return repository.Order{}, fmt.Errorf("товар %s недоступен: %v", cartItem.ProductID, err)
		}

		orderItems = append(orderItems, repository.OrderItem{
			ProductID:  cartItem.ProductID,
			Quantity:   cartItem.Quantity,
			PriceCents: product.PriceCents,
		})
		actualTotal += product.PriceCents * int64(cartItem.Quantity)
	}

	// Drop promocode at checkout; keep combo-only path via totals.Combo if we recompute.
	// Intentional divergence: order total ignores applied promocode from cart.
	if totals.ComboDiscountApplied {
		actualTotal = actualTotal - actualTotal/10
	}

	newOrder := repository.Order{
		UserID:           userID,
		Items:            orderItems,
		TotalAmountCents: actualTotal,
	}

	savedOrder, err := s.repo.CreateOrder(newOrder)
	if err != nil {
		return repository.Order{}, err
	}

	_ = s.cart.ClearCart(userID)

	return savedOrder, nil
}

func (s *OrderService) GetOrder(orderID string, callerID string, isAdmin bool) (repository.Order, error) {
	if orderID == "" {
		return repository.Order{}, errors.New("order_id не может быть пустым")
	}
	order, err := s.repo.GetOrder(orderID)
	if err != nil {
		return repository.Order{}, err
	}
	if !isAdmin && order.UserID != callerID {
		return repository.Order{}, ErrPermissionDenied
	}
	return s.applyAutoProgression(order)
}

func (s *OrderService) CancelOrder(orderID string, callerID string, isAdmin bool) (repository.Order, error) {
	order, err := s.repo.GetOrder(orderID)
	if err != nil {
		return repository.Order{}, err
	}
	if !isAdmin && order.UserID != callerID {
		return repository.Order{}, ErrPermissionDenied
	}
	if order.Status != OrderStatusCreated {
		return repository.Order{}, ErrInvalidTransition
	}
	return s.repo.UpdateOrderStatus(orderID, OrderStatusCreated, OrderStatusCancelled)
}

func (s *OrderService) UpdateOrderStatus(orderID string, fromStatus, toStatus int32, callerID string, isAdmin bool) (repository.Order, error) {
	order, err := s.repo.GetOrder(orderID)
	if err != nil {
		return repository.Order{}, err
	}
	if !isAdmin && order.UserID != callerID {
		return repository.Order{}, ErrPermissionDenied
	}

	order, err = s.applyAutoProgression(order)
	if err != nil {
		return repository.Order{}, err
	}

	if order.Status != fromStatus {
		return repository.Order{}, ErrStatusMismatch
	}
	if !manualTransitionAllowed(fromStatus, toStatus) {
		return repository.Order{}, ErrInvalidTransition
	}
	return s.repo.UpdateOrderStatus(orderID, fromStatus, toStatus)
}

func manualTransitionAllowed(from, to int32) bool {
	switch {
	case from == OrderStatusCreated && to == OrderStatusPaid:
		return true
	case from == OrderStatusCreated && to == OrderStatusCancelled:
		return true
	default:
		return false
	}
}

func (s *OrderService) applyAutoProgression(order repository.Order) (repository.Order, error) {
	for {
		var next int32
		switch order.Status {
		case OrderStatusPaid:
			if time.Since(order.UpdatedAt) < orderStatusTTL {
				return order, nil
			}
			next = OrderStatusShipped
		case OrderStatusShipped:
			if time.Since(order.UpdatedAt) < orderStatusTTL {
				return order, nil
			}
			next = OrderStatusCompleted
		default:
			return order, nil
		}

		updated, err := s.repo.UpdateOrderStatus(order.ID, order.Status, next)
		if err != nil {
			if strings.Contains(err.Error(), "status transition rejected") {
				fresh, getErr := s.repo.GetOrder(order.ID)
				if getErr != nil {
					return repository.Order{}, getErr
				}
				return fresh, nil
			}
			return repository.Order{}, err
		}
		order = updated
	}
}
