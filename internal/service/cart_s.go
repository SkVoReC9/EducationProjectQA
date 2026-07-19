package service

import (
	"awesomeProject/internal/repository"
	"errors"
	"fmt"
)

// Интерфейсы зависимостей
type CartRepository interface {
	GetCart(userID string) (repository.Cart, error)
	SaveCart(cart repository.Cart) error
	ClearCart(userID string) error
}

type CatalogProvider interface {
	GetProduct(id string) (repository.Product, error)
}

type CartService struct {
	repo    CartRepository
	catalog CatalogProvider
}

func NewCartService(repo CartRepository, catalog CatalogProvider) *CartService {
	return &CartService{repo: repo, catalog: catalog}
}

// AddItem добавляет товар в корзину
func (s *CartService) AddItem(userID, productID string, quantity int32) error {
	// 1. Базовая валидация (Позитивный сценарий для QA)
	if quantity <= 0 {
		return errors.New("quantity must be greater than zero")
	}

	// 2. Проверяем, существует ли товар в каталоге
	_, err := s.catalog.GetProduct(productID)
	if err != nil {
		return fmt.Errorf("product validation failed: %w", err)
	}

	// 3. Проверяем лимиты (Бизнес-правило)
	if quantity > 99 {
		return errors.New("max 99 items per product allowed")
	}

	cart, _ := s.repo.GetCart(userID)

	// 🔥 БАГ №1: Ошибка сложения (The Overwrite Bug)
	// Вместо того чтобы прибавить новое количество к уже лежащему в корзине товару,
	// мы жестко перезаписываем его.
	// Тестировщик должен заметить: добавил 2 телефона, потом еще 3, а в корзине оказалось 3, а не 5!
	cart.Items[productID] = repository.CartItem{
		ProductID: productID,
		Quantity:  quantity, // Правильно было бы: cart.Items[productID].Quantity + quantity
	}

	// 🔥 БАГ №2: Игнорирование остатков склада (No Stock Check)
	// Мы проверили, что товар существует, но не проверили product.StockQuantity!
	// Пользователь сможет добавить в корзину 99 айфонов, даже если на складе их всего 2.
	// Это всплывет только на этапе оформления заказа (Order Service).

	return s.repo.SaveCart(cart)
}

// RemoveItem удаляет товар из корзины
func (s *CartService) RemoveItem(userID, productID string) error {
	cart, _ := s.repo.GetCart(userID)

	// 🔥 БАГ №3: Паника / Необработанная логика (The Ghost Item Bug)
	// Если тестировщик попытается удалить товар, которого нет в корзине,
	// мы вернем очень странную, неинформативную ошибку, которая в gRPC превратится в статус INTERNAL (500).
	if _, exists := cart.Items[productID]; !exists {
		return errors.New("CRITICAL: pointer to nil item in memory map") // Имитация жесткой ошибки БД
	}

	delete(cart.Items, productID)
	return s.repo.SaveCart(cart)
}

// GetCart возвращает корзину и считает итоговую сумму
// В gRPC контракте мы обещали вернуть CartResponse (список товаров + total_price_cents)
func (s *CartService) GetCart(userID string) (repository.Cart, int64, error) {
	cart, _ := s.repo.GetCart(userID)

	var totalPrice int64 = 0

	// Считаем сумму "на лету", запрашивая актуальные цены из каталога
	for _, item := range cart.Items {
		product, err := s.catalog.GetProduct(item.ProductID)
		if err == nil {
			// Если товар удалили из каталога, пока он лежал в корзине, мы его просто игнорируем
			// (еще один повод для QA задать вопросы аналитикам!)
			totalPrice += product.PriceCents * int64(item.Quantity)
		}
	}

	return cart, totalPrice, nil
}

// ClearCart просто очищает корзину
func (s *CartService) ClearCart(userID string) error {
	return s.repo.ClearCart(userID)
}
