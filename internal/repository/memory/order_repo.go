package memory

import (
	"fmt"
	"sync"
	"time"

	"awesomeProject/internal/repository"
)

type OrderRepository struct {
	mu     sync.RWMutex
	orders map[string]repository.Order
}

func NewOrderRepository() *OrderRepository {
	return &OrderRepository{
		orders: make(map[string]repository.Order),
	}
}

func (r *OrderRepository) CreateOrder(order repository.Order) (repository.Order, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Генерируем простейший уникальный ID на основе времени
	order.ID = fmt.Sprintf("ord-%d", time.Now().UnixMilli())
	order.Status = 1 // ORDER_STATUS_CREATED

	r.orders[order.ID] = order
	return order, nil
}

func (r *OrderRepository) GetOrder(orderID string) (repository.Order, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	order, exists := r.orders[orderID]
	if !exists {
		return repository.Order{}, fmt.Errorf("заказ не найден")
	}
	return order, nil
}
