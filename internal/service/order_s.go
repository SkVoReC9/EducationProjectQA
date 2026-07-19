package service

import (
	"errors"
	"fmt"

	"awesomeProject/internal/repository"
)

type OrderRepository interface {
	CreateOrder(order repository.Order) (repository.Order, error)
	GetOrder(orderID string) (repository.Order, error)
}

// Нам нужен доступ к корзине, чтобы забрать оттуда товары
type CartProvider interface {
	GetCart(userID string) (repository.Cart, int64, error)
	ClearCart(userID string) error
}

type OrderService struct {
	repo    OrderRepository
	cart    CartProvider
	catalog CatalogProvider // Используем уже существующий интерфейс из cart_s.go
}

func NewOrderService(repo OrderRepository, cart CartProvider, catalog CatalogProvider) *OrderService {
	return &OrderService{repo: repo, cart: cart, catalog: catalog}
}

func (s *OrderService) CreateOrder(userID string) (repository.Order, error) {
	// 1. Получаем корзину пользователя
	cart, expectedTotal, err := s.cart.GetCart(userID)
	if err != nil {
		return repository.Order{}, fmt.Errorf("не удалось получить корзину: %v", err)
	}

	// 🔥 БАГ №4: Оформление воздуха (The Ghost Order Bug)
	// Мы забыли проверить, а есть ли вообще товары в корзине!
	// Тестировщик сможет отправить запрос с пустым user_id или очищенной корзиной
	// и успешно создать заказ на сумму 0 копеек с нулем товаров.
	/* Правильно было бы написать:
	if len(cart.Items) == 0 {
		return repository.Order{}, errors.New("корзина пуста")
	}
	*/

	var orderItems []repository.OrderItem
	var actualTotal int64 = 0

	// 2. Формируем позиции заказа
	for _, cartItem := range cart.Items {
		// Снова запрашиваем каталог, чтобы зафиксировать актуальную цену
		product, err := s.catalog.GetProduct(cartItem.ProductID)
		if err != nil {
			return repository.Order{}, fmt.Errorf("товар %s недоступен: %v", cartItem.ProductID, err)
		}

		// 🔥 Здесь всплывает БАГ №2 из корзины (No Stock Check)!
		// Мы по-прежнему нигде не проверяем product.StockQuantity.
		// Пользователь купит 100 айфонов, даже если на складе их 10.

		orderItems = append(orderItems, repository.OrderItem{
			ProductID:  cartItem.ProductID,
			Quantity:   cartItem.Quantity,
			PriceCents: product.PriceCents,
		})
		actualTotal += product.PriceCents * int64(cartItem.Quantity)
	}

	// 3. Создаем объект заказа
	newOrder := repository.Order{
		UserID:           userID,
		Items:            orderItems,
		TotalAmountCents: actualTotal,
	}

	// 4. Сохраняем заказ в базу
	savedOrder, err := s.repo.CreateOrder(newOrder)
	if err != nil {
		return repository.Order{}, err
	}

	// 5. Очищаем корзину после успешной покупки
	_ = s.cart.ClearCart(userID)

	// Мы игнорируем expectedTotal из корзины. Если цена поменялась
	// прямо во время оформления, пользователь узнает об этом только по факту.
	_ = expectedTotal

	return savedOrder, nil
}

func (s *OrderService) GetOrder(orderID string) (repository.Order, error) {
	if orderID == "" {
		return repository.Order{}, errors.New("order_id не может быть пустым")
	}
	return s.repo.GetOrder(orderID)
}
