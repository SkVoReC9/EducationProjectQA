# Документация API — Store API Simulator

Симулятор магазина с двумя транспортами: **REST/JSON** (HTTP-шлюз) и **gRPC**. Данные хранятся в **PostgreSQL**.

| Порт  | Протокол                         |
|-------|----------------------------------|
| 8080  | HTTP REST (JSON, grpc-gateway)   |
| 50051 | gRPC                             |

Базовый URL REST: `http://localhost:8080`

---

## Общие сведения

### Seed-каталог

50 товаров: **15 Apple**, **15 Samsung**, **10 NVIDIA**, **10 AMD**.

| Бренд   | Диапазон UUID (суффикс) | Примеры |
|---------|-------------------------|---------|
| Apple   | `...001`–`...015` | iPhone 15, MacBook Pro 14 M3, AirPods Pro 2 |
| Samsung | `...016`–`...030` | Galaxy S24 Ultra, Odyssey G9, 990 PRO |
| NVIDIA  | `...031`–`...040` | RTX 4090, RTX 4070 SUPER, Jetson Orin Nano |
| AMD     | `...041`–`...050` | Ryzen 9 7950X, RX 7900 XTX, EPYC 9654 |

Префикс UUID: `550e8400-e29b-41d4-a716-44665544` + `0001`…`0050`. Полный список — в `migrations/002_seed.sql`.

Пример ID iPhone 15: `550e8400-e29b-41d4-a716-446655440001`.

### Цены

Все денежные поля передаются в **центах** (`price_cents`, `total_price_cents`, `total_amount_cents`). Например, `99900` = $999.00.

### Ошибки

| gRPC-код              | HTTP-код | Когда возникает |
|-----------------------|----------|-----------------|
| `InvalidArgument`     | 400      | Пустые обязательные поля, невалидное количество |
| `NotFound`            | 404      | Товар или заказ не найден |
| `FailedPrecondition`  | 400      | Нарушение предусловия (например, пустая корзина — если проверка включена) |
| `Internal`            | 500      | Внутренняя ошибка сервера |

Через REST ошибки возвращаются в формате gRPC-gateway (JSON с полями `code`, `message`, `details`).

---

# REST API

## 1. Каталог

### 1.1. Список товаров

```http
GET /v1/products
```

**Query-параметры** (зарезервированы, пока не используются):

| Параметр     | Тип    | Описание              |
|--------------|--------|-----------------------|
| `page_size`  | int32  | Размер страницы       |
| `page_token` | string | Токен следующей страницы |

**Ответ `200`:**

```json
{
  "products": [
    {
      "id": "550e8400-e29b-41d4-a716-446655440001",
      "name": "iPhone 15",
      "description": "Apple smartphone with A16 Bionic",
      "priceCents": "79900",
      "stockQuantity": 40
    },
    {
      "id": "550e8400-e29b-41d4-a716-446655440002",
      "name": "iPhone 15 Pro",
      "description": "Apple flagship with titanium frame",
      "priceCents": "99900",
      "stockQuantity": 25
    },
    {
      "id": "550e8400-e29b-41d4-a716-446655440016",
      "name": "Galaxy S24",
      "description": "Samsung flagship smartphone",
      "priceCents": "79900",
      "stockQuantity": 35
    }
  ],
  "nextPageToken": ""
}
```

**Пример:**

```bash
curl http://localhost:8080/v1/products
```

---

### 1.2. Получить товар по ID

```http
GET /v1/products/{product_id}
```

| Параметр пути | Тип    | Обязательный | Описание   |
|---------------|--------|--------------|------------|
| `product_id`  | string | да           | ID товара  |

**Ответ `200`:**

```json
{
  "product": {
    "id": "550e8400-e29b-41d4-a716-446655440001",
    "name": "iPhone 15",
    "description": "Apple smartphone with A16 Bionic",
    "priceCents": "79900",
    "stockQuantity": 40
  }
}
```

**Ошибки:**

| Условие              | HTTP | Сообщение |
|----------------------|------|-----------|
| Товар не найден      | 404  | Товар с ID '...' не найден |
| Пустой `product_id`  | 400  | ID товара не может быть пустым |

**Пример:**

```bash
curl http://localhost:8080/v1/products/550e8400-e29b-41d4-a716-446655440001
```

---

## 2. Корзина

### 2.1. Получить корзину

```http
GET /v1/users/{user_id}/cart
```

| Параметр пути | Тип    | Обязательный | Описание        |
|---------------|--------|--------------|-----------------|
| `user_id`     | string | да           | ID пользователя |

Если корзины ещё нет, возвращается пустой список товаров и сумма `0`.

**Ответ `200`:**

```json
{
  "items": [
    {
      "productId": "550e8400-e29b-41d4-a716-446655440001",
      "quantity": 2
    }
  ],
  "totalPriceCents": "159800"
}
```

**Ошибки:**

| Условие         | HTTP | Сообщение            |
|-----------------|------|----------------------|
| Пустой `user_id`| 400  | user_id обязателен   |

**Пример:**

```bash
curl http://localhost:8080/v1/users/user-1/cart
```

---

### 2.2. Добавить товар в корзину

```http
POST /v1/users/{user_id}/cart/items
Content-Type: application/json
```

| Параметр пути | Тип    | Обязательный | Описание        |
|---------------|--------|--------------|-----------------|
| `user_id`     | string | да           | ID пользователя |

**Тело запроса:**

```json
{
  "product_id": "550e8400-e29b-41d4-a716-446655440001",
  "quantity": 1
}
```

| Поле         | Тип    | Обязательный | Описание                          |
|--------------|--------|--------------|-----------------------------------|
| `product_id` | string | да           | ID товара из каталога             |
| `quantity`   | int32  | да           | Количество (от 1 до 99 включительно) |

**Ответ `200`:** актуальная корзина (`items` + `totalPriceCents`).

**Ошибки:**

| Условие                         | HTTP | Сообщение |
|---------------------------------|------|-----------|
| Пустой `user_id`                | 400  | user_id обязателен |
| `quantity <= 0`                 | 400  | quantity must be greater than zero |
| `quantity > 99`                 | 400  | max 99 items per product allowed |
| Товар не найден в каталоге      | 404  | товар не найден в каталоге |

**Пример:**

```bash
curl -X POST http://localhost:8080/v1/users/user-1/cart/items \
  -H 'Content-Type: application/json' \
  -d '{"product_id":"550e8400-e29b-41d4-a716-446655440001","quantity":1}'
```

---

### 2.3. Удалить товар из корзины

```http
DELETE /v1/users/{user_id}/cart/items/{product_id}
```

| Параметр пути | Тип    | Обязательный | Описание        |
|---------------|--------|--------------|-----------------|
| `user_id`     | string | да           | ID пользователя |
| `product_id`  | string | да           | ID товара       |

**Ответ `200`:** актуальная корзина после удаления.

**Ошибки:**

| Условие                              | HTTP | Сообщение |
|--------------------------------------|------|-----------|
| Пустой `user_id`                     | 400  | user_id обязателен |
| Товара нет в корзине                 | 500  | CRITICAL: pointer to nil item in memory map |

**Пример:**

```bash
curl -X DELETE http://localhost:8080/v1/users/user-1/cart/items/550e8400-e29b-41d4-a716-446655440001
```

---

### 2.4. Очистить корзину

```http
DELETE /v1/users/{user_id}/cart
```

| Параметр пути | Тип    | Обязательный | Описание        |
|---------------|--------|--------------|-----------------|
| `user_id`     | string | да           | ID пользователя |

**Ответ `200`:** пустая корзина.

```json
{
  "items": [],
  "totalPriceCents": "0"
}
```

**Ошибки:**

| Условие          | HTTP | Сообщение          |
|------------------|------|--------------------|
| Пустой `user_id` | 400  | user_id обязателен |

**Пример:**

```bash
curl -X DELETE http://localhost:8080/v1/users/user-1/cart
```

---

## 3. Заказы

### 3.1. Создать заказ из корзины

```http
POST /v1/orders
Content-Type: application/json
```

Создаёт заказ на основе текущей корзины пользователя. После успешного создания корзина очищается. Статус нового заказа — `ORDER_STATUS_CREATED` (1).

**Тело запроса:**

```json
{
  "user_id": "user-1"
}
```

| Поле      | Тип    | Обязательный | Описание        |
|-----------|--------|--------------|-----------------|
| `user_id` | string | да           | ID пользователя |

**Ответ `200`:**

```json
{
  "order": {
    "id": "a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11",
    "userId": "user-1",
    "items": [
      {
        "productId": "550e8400-e29b-41d4-a716-446655440001",
        "quantity": 1,
        "priceCents": "79900"
      }
    ],
    "totalAmountCents": "79900",
    "status": "ORDER_STATUS_CREATED"
  }
}
```

**Ошибки:**

| Условие                    | HTTP | Сообщение |
|----------------------------|------|-----------|
| Пустой `user_id`           | 400  | user_id обязателен |
| Товар из корзины недоступен| 500  | ошибка создания заказа: ... |

**Пример:**

```bash
curl -X POST http://localhost:8080/v1/orders \
  -H 'Content-Type: application/json' \
  -d '{"user_id":"user-1"}'
```

---

### 3.2. Получить заказ по ID

```http
GET /v1/orders/{order_id}
```

| Параметр пути | Тип    | Обязательный | Описание  |
|---------------|--------|--------------|-----------|
| `order_id`    | string | да           | ID заказа |

**Ответ `200`:** объект `order` (см. выше).

**Ошибки:**

| Условие           | HTTP | Сообщение          |
|-------------------|------|--------------------|
| Пустой `order_id` | 400  | order_id обязателен |
| Заказ не найден   | 404  | заказ не найден    |

**Пример:**

```bash
curl http://localhost:8080/v1/orders/a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11
```

---

### Статусы заказа (REST / JSON)

| Значение JSON              | Число | Описание              |
|----------------------------|-------|-----------------------|
| `ORDER_STATUS_UNSPECIFIED` | 0     | Не задан              |
| `ORDER_STATUS_CREATED`     | 1     | Создан, ожидает оплаты|
| `ORDER_STATUS_PAID`        | 2     | Оплачен               |
| `ORDER_STATUS_SHIPPED`     | 3     | Отправлен             |
| `ORDER_STATUS_CANCELLED`   | 4     | Отменён               |

---

## Пример полного сценария (REST)

```bash
# 1. Список товаров
curl http://localhost:8080/v1/products

# 2. Добавить в корзину
curl -X POST http://localhost:8080/v1/users/user-1/cart/items \
  -H 'Content-Type: application/json' \
  -d '{"product_id":"550e8400-e29b-41d4-a716-446655440001","quantity":1}'

# 3. Посмотреть корзину
curl http://localhost:8080/v1/users/user-1/cart

# 4. Оформить заказ
curl -X POST http://localhost:8080/v1/orders \
  -H 'Content-Type: application/json' \
  -d '{"user_id":"user-1"}'

# 5. Получить заказ (подставьте id из ответа шага 4)
curl http://localhost:8080/v1/orders/<order-uuid>
```

---

# gRPC API

Адрес: `localhost:50051`  
Reflection включён — удобно смотреть методы через `grpcurl` / Postman / BloomRPC.

Контракты: `proto/catalog.proto`, `proto/cart.proto`, `proto/order.proto`.

## Сервисы

| Сервис                         | Пакет              | Описание        |
|--------------------------------|--------------------|-----------------|
| `CatalogService`               | `store.catalog.v1` | Каталог товаров |
| `CartService`                  | `store.cart.v1`    | Корзина         |
| `OrderService`                 | `store.order.v1`   | Заказы          |

---

## 1. CatalogService (`store.catalog.v1.CatalogService`)

### 1.1. ListProducts

Список товаров.

**Запрос `ListProductsRequest`:**

| Поле         | Тип    | Описание                         |
|--------------|--------|----------------------------------|
| `page_size`  | int32  | Зарезервировано (не используется)|
| `page_token` | string | Зарезервировано (не используется)|

**Ответ `ListProductsResponse`:**

| Поле              | Тип               | Описание                    |
|-------------------|-------------------|-----------------------------|
| `products`        | repeated Product  | Список товаров              |
| `next_page_token` | string            | Пока всегда пустая строка   |

**Модель `Product`:**

| Поле             | Тип    | Описание                |
|------------------|--------|-------------------------|
| `id`             | string | ID товара               |
| `name`           | string | Название                |
| `description`    | string | Описание                |
| `price_cents`    | int64  | Цена в центах           |
| `stock_quantity` | int32  | Остаток на складе       |

**Пример (grpcurl):**

```bash
grpcurl -plaintext localhost:50051 store.catalog.v1.CatalogService/ListProducts
```

---

### 1.2. GetProduct

Получить товар по ID.

**Запрос `GetProductRequest`:**

| Поле         | Тип    | Обязательный | Описание  |
|--------------|--------|--------------|-----------|
| `product_id` | string | да           | ID товара |

**Ответ `ProductResponse`:**

| Поле      | Тип     | Описание |
|-----------|---------|----------|
| `product` | Product | Товар    |

**Ошибки:** `InvalidArgument`, `NotFound`, `Internal`.

**Пример:**

```bash
grpcurl -plaintext -d '{"product_id":"550e8400-e29b-41d4-a716-446655440001"}' \
  localhost:50051 store.catalog.v1.CatalogService/GetProduct
```

---

## 2. CartService (`store.cart.v1.CartService`)

### Общий ответ `CartResponse`

| Поле                | Тип                 | Описание                    |
|---------------------|---------------------|-----------------------------|
| `items`             | repeated CartItem   | Позиции корзины             |
| `total_price_cents` | int64               | Итоговая сумма в центах     |

**Модель `CartItem`:**

| Поле         | Тип    | Описание     |
|--------------|--------|--------------|
| `product_id` | string | ID товара    |
| `quantity`   | int32  | Количество |

---

### 2.1. GetCart

**Запрос `GetCartRequest`:**

| Поле      | Тип    | Обязательный |
|-----------|--------|--------------|
| `user_id` | string | да           |

**Пример:**

```bash
grpcurl -plaintext -d '{"user_id":"user-1"}' \
  localhost:50051 store.cart.v1.CartService/GetCart
```

---

### 2.2. AddItem

**Запрос `AddItemRequest`:**

| Поле         | Тип    | Обязательный | Описание        |
|--------------|--------|--------------|-----------------|
| `user_id`    | string | да           | ID пользователя |
| `product_id` | string | да           | ID товара       |
| `quantity`   | int32  | да           | 1–99            |

**Ошибки:** `InvalidArgument`, `NotFound`, `Internal`.

**Пример:**

```bash
grpcurl -plaintext -d '{"user_id":"user-1","product_id":"550e8400-e29b-41d4-a716-446655440001","quantity":1}' \
  localhost:50051 store.cart.v1.CartService/AddItem
```

---

### 2.3. RemoveItem

**Запрос `RemoveItemRequest`:**

| Поле         | Тип    | Обязательный |
|--------------|--------|--------------|
| `user_id`    | string | да           |
| `product_id` | string | да           |

**Ошибки:** `InvalidArgument`, `Internal` (если товара нет в корзине).

**Пример:**

```bash
grpcurl -plaintext -d '{"user_id":"user-1","product_id":"550e8400-e29b-41d4-a716-446655440001"}' \
  localhost:50051 store.cart.v1.CartService/RemoveItem
```

---

### 2.4. ClearCart

**Запрос `ClearCartRequest`:**

| Поле      | Тип    | Обязательный |
|-----------|--------|--------------|
| `user_id` | string | да           |

**Пример:**

```bash
grpcurl -plaintext -d '{"user_id":"user-1"}' \
  localhost:50051 store.cart.v1.CartService/ClearCart
```

---

## 3. OrderService (`store.order.v1.OrderService`)

### Модель `Order`

| Поле                 | Тип                 | Описание                    |
|----------------------|---------------------|-----------------------------|
| `id`                 | string              | ID заказа (`UUID`)|
| `user_id`            | string              | ID пользователя             |
| `items`              | repeated OrderItem  | Позиции заказа              |
| `total_amount_cents` | int64               | Сумма заказа в центах       |
| `status`             | OrderStatus         | Статус                      |

**Модель `OrderItem`:**

| Поле          | Тип    | Описание                          |
|---------------|--------|-----------------------------------|
| `product_id`  | string | ID товара                         |
| `quantity`    | int32  | Количество                       |
| `price_cents` | int64  | Цена на момент оформления заказа  |

**Enum `OrderStatus`:**

| Имя                        | Значение | Описание               |
|----------------------------|----------|------------------------|
| `ORDER_STATUS_UNSPECIFIED` | 0        | Не задан               |
| `ORDER_STATUS_CREATED`     | 1        | Создан, ожидает оплаты |
| `ORDER_STATUS_PAID`        | 2        | Оплачен                |
| `ORDER_STATUS_SHIPPED`     | 3        | Отправлен              |
| `ORDER_STATUS_CANCELLED`   | 4        | Отменён                |

---

### 3.1. CreateOrder

Создаёт заказ из корзины пользователя и очищает корзину.

**Запрос `CreateOrderRequest`:**

| Поле      | Тип    | Обязательный |
|-----------|--------|--------------|
| `user_id` | string | да           |

**Ответ `OrderResponse`:**

| Поле    | Тип   | Описание |
|---------|-------|----------|
| `order` | Order | Заказ    |

**Ошибки:** `InvalidArgument`, `FailedPrecondition`, `Internal`.

**Пример:**

```bash
grpcurl -plaintext -d '{"user_id":"user-1"}' \
  localhost:50051 store.order.v1.OrderService/CreateOrder
```

---

### 3.2. GetOrder

**Запрос `GetOrderRequest`:**

| Поле       | Тип    | Обязательный |
|------------|--------|--------------|
| `order_id` | string | да           |

**Ошибки:** `InvalidArgument`, `NotFound`.

**Пример:**

```bash
grpcurl -plaintext -d '{"order_id":"a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11"}' \
  localhost:50051 store.order.v1.OrderService/GetOrder
```

---

## Соответствие REST ↔ gRPC

| REST                                      | gRPC-метод                              |
|-------------------------------------------|-----------------------------------------|
| `GET /v1/products`                        | `CatalogService.ListProducts`           |
| `GET /v1/products/{product_id}`           | `CatalogService.GetProduct`             |
| `GET /v1/users/{user_id}/cart`            | `CartService.GetCart`                   |
| `POST /v1/users/{user_id}/cart/items`     | `CartService.AddItem`                   |
| `DELETE /v1/users/{user_id}/cart/items/{product_id}` | `CartService.RemoveItem`     |
| `DELETE /v1/users/{user_id}/cart`         | `CartService.ClearCart`                 |
| `POST /v1/orders`                         | `OrderService.CreateOrder`              |
| `GET /v1/orders/{order_id}`               | `OrderService.GetOrder`                 |

---

## Полезные команды grpcurl

```bash
# Список сервисов
grpcurl -plaintext localhost:50051 list

# Описание методов сервиса
grpcurl -plaintext localhost:50051 describe store.cart.v1.CartService

# Описание сообщения
grpcurl -plaintext localhost:50051 describe store.order.v1.Order
```
