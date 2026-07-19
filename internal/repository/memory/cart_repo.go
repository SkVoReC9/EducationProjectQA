package memory

import (
	"sync"

	// Замени "store" на имя твоего модуля
	"awesomeProject/internal/repository"
)

// CartRepository имитирует базу данных для корзин
type CartRepository struct {
	mu    sync.RWMutex
	carts map[string]repository.Cart // ключ - user_id
}

func NewCartRepository() *CartRepository {
	return &CartRepository{
		carts: make(map[string]repository.Cart),
	}
}

// GetCart возвращает корзину пользователя.
// В отличие от каталога, если корзины нет, мы не возвращаем ошибку, а отдаем пустую корзину.
func (r *CartRepository) GetCart(userID string) (repository.Cart, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	cart, exists := r.carts[userID]
	if !exists {
		// Инициализируем новую корзину "на лету"
		return repository.Cart{
			UserID: userID,
			Items:  make(map[string]repository.CartItem),
		}, nil
	}

	// Важно: возвращаем копию мапы, чтобы избежать race conditions,
	// если сервис попытается изменить данные напрямую без вызова SaveCart
	cartCopy := repository.Cart{
		UserID: cart.UserID,
		Items:  make(map[string]repository.CartItem),
	}
	for k, v := range cart.Items {
		cartCopy.Items[k] = v
	}

	return cartCopy, nil
}

// SaveCart перезаписывает состояние корзины в базе
func (r *CartRepository) SaveCart(cart repository.Cart) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.carts[cart.UserID] = cart
	return nil
}

// ClearCart полностью удаляет корзину пользователя (пригодится при оформлении заказа)
func (r *CartRepository) ClearCart(userID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	delete(r.carts, userID)
	return nil
}
