package repository

// --- Модели Каталога ---

// Product - внутренняя модель товара
type Product struct {
	ID            string
	Name          string
	Description   string
	PriceCents    int64
	StockQuantity int32
}

// --- Модели Корзины ---

// CartItem - позиция в корзине
type CartItem struct {
	ProductID string
	Quantity  int32
}

// Cart - корзина конкретного пользователя
type Cart struct {
	UserID string
	Items  map[string]CartItem // Ключ - ProductID для быстрого поиска
}

// OrderItem - зафиксированная позиция в чеке
type OrderItem struct {
	ProductID  string
	Quantity   int32
	PriceCents int64 // Цена фиксируется на момент покупки!
}

// Order - итоговый заказ пользователя
type Order struct {
	ID               string
	UserID           string
	Items            []OrderItem
	TotalAmountCents int64
	Status           int32 // 1 - CREATED, 2 - PAID, etc.
}
