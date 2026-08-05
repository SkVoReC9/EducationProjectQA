# Документация API — Store API Simulator

Симулятор магазина с двумя транспортами: **REST/JSON** (HTTP-шлюз) и **gRPC**. Данные хранятся в **PostgreSQL**.

История обновлений по итерациям: [`RELEASE.md`](RELEASE.md).

| Порт  | Протокол                         |
|-------|----------------------------------|
| 8080  | HTTP REST (JSON, grpc-gateway)   |
| 50051 | gRPC                             |

Базовый URL REST: `http://localhost:8080`

JSON от gateway — **camelCase** (`userId`, `accessToken`, `totalPriceCents`, …).

---

## Общие сведения

### Auth

| Область | Требование |
|---------|------------|
| Каталог, register, login, get user | Публично |
| Корзина, заказы | `Authorization: Bearer <access_token>` |
| Admin-промокоды `/v1/admin/*` | JWT с ролью `admin` |

- TTL токена: **24 часа**.
- В JWT: `sub` = user id, `role` = `user` \| `admin`.
- На cart/order `user_id` в пути/теле **должен совпадать** с `sub` → иначе **403**.
- Без токена / битый токен на protected → **401**.
- Не-admin на admin API → **403**.

Сидовый админ:

| Поле | Значение |
|------|----------|
| Email | `admin@store.local` |
| Password | `admin123` |
| Role | `admin` |

### Seed-каталог

50 товаров: **15 Apple**, **15 Samsung**, **10 NVIDIA**, **10 AMD**. У каждого товара есть поле `brand`: `apple` / `samsung` / `nvidia` / `amd`.

| Бренд   | Диапазон UUID (суффикс) | Примеры |
|---------|-------------------------|---------|
| Apple   | `...001`–`...015` | iPhone 15, MacBook Pro 14 M3, AirPods Pro 2 |
| Samsung | `...016`–`...030` | Galaxy S24 Ultra, Odyssey G9, 990 PRO |
| NVIDIA  | `...031`–`...040` | RTX 4090, RTX 4070 SUPER, Jetson Orin Nano |
| AMD     | `...041`–`...050` | Ryzen 9 7950X, RX 7900 XTX, EPYC 9654 |

Префикс UUID: `550e8400-e29b-41d4-a716-44665544` + `0001`…`0050`. Полный список — в `migrations/002_seed.sql`.

Пример ID iPhone 15: `550e8400-e29b-41d4-a716-446655440001`.

### Цены и скидки (ожидаемое поведение)

Все денежные поля — в **центах**.

**Сабтотал** — сумма `priceCents * quantity` по позициям корзины.

**Combo-скидка 10%** — если в корзине есть:

- ≥1 товар с `brand = nvidia`, **и**
- ≥1 товар, в **названии** которого есть `iPhone`

Тогда 10% от сабтотала. MacBook / AirPods / прочий Apple **без** `iPhone` в имени combo **не** активируют.

**Промокод** — применяется отдельно (`POST .../cart/promocode`).

Сидовые коды:

| Код | Эффект |
|-----|--------|
| `SAVE10` | 10% |
| `FLAT500` | 500 центов |
| `WELCOME` | 15% |

Ожидаемые правила промо:

1. Коды **без учёта регистра** (`save10` ≡ `SAVE10`).
2. Промокод и combo **не суммируются**: если оба применимы, действует **только промокод**.
3. `ClearCart` очищает позиции **и** снимает промокод.
4. Итог заказа при `POST /v1/orders` должен совпадать с итогом `GET` корзины (те же скидки).

**TTL корзины:** 30 минут неактивности. Обновляется при add / remove / clear / apply/clear promocode. **Не** обновляется от простого `GET` корзины. После истечения — пустая корзина.

### Ошибки

| gRPC-код              | HTTP-код | Когда возникает |
|-----------------------|----------|-----------------|
| `InvalidArgument`     | 400      | Пустые обязательные поля, невалидное количество |
| `Unauthenticated`     | 401      | Нет / невалидный JWT |
| `PermissionDenied`    | 403      | Чужой `user_id` или не admin |
| `NotFound`            | 404      | Товар, заказ, пользователь, промокод не найден |
| `FailedPrecondition`  | 400      | Недопустимый переход статуса, пустая корзина (если проверка включена) |
| `Internal`            | 500      | Внутренняя ошибка сервера |

Через REST ошибки — формат gRPC-gateway (`code`, `message`, `details`).

---

# REST API

## 0. Пользователи

### 0.1. Регистрация

```http
POST /v1/users/register
Content-Type: application/json
```

```json
{ "email": "qa@example.com", "password": "secret123", "name": "QA" }
```

**Ответ `200`:** `user` + `accessToken`. У `user` есть `id`, `email`, `name`, `role`, `createdAt`.

### 0.2. Вход

```http
POST /v1/users/login
```

```json
{ "email": "qa@example.com", "password": "secret123" }
```

**Ответ `200`:** `user` + `accessToken`. Неверные данные → **401**.

### 0.3. Получить пользователя

```http
GET /v1/users/{user_id}
```

### 0.4. Удалить пользователя

```http
DELETE /v1/users/{user_id}
Authorization: Bearer <access_token>
```

Требуется JWT; `user_id` в пути должен совпадать с `sub` токена.
Заказы в `COMPLETED` / `CANCELLED` удаляются вместе с пользователем.
Если есть активные заказы (`CREATED` / `PAID` / `SHIPPED`) → **FailedPrecondition**.

---

## 1. Каталог

### 1.1. Список товаров

```http
GET /v1/products
```

**Ответ `200`:**

```json
{
  "products": [
    {
      "id": "550e8400-e29b-41d4-a716-446655440001",
      "name": "iPhone 15",
      "description": "Apple smartphone with A16 Bionic",
      "priceCents": "79900",
      "stockQuantity": 40,
      "brand": "apple"
    }
  ],
  "nextPageToken": ""
}
```

```bash
curl http://localhost:8080/v1/products
```

---

### 1.2. Получить товар по ID

```http
GET /v1/products/{product_id}
```

**Ошибки:** товар не найден → **404**; пустой id → **400**.

```bash
curl http://localhost:8080/v1/products/550e8400-e29b-41d4-a716-446655440001
```

---

## 2. Корзина (JWT)

Во все запросы: `-H "Authorization: Bearer ${TOKEN}"`.

### Ответ корзины

```json
{
  "items": [{ "productId": "...", "quantity": 1 }],
  "totalPriceCents": "71910",
  "subtotalCents": "79900",
  "discountCents": "7990",
  "appliedPromocode": "SAVE10",
  "comboDiscountApplied": false,
  "expiresAt": "2026-07-26T12:30:00Z"
}
```

| Поле | Описание |
|------|----------|
| `subtotalCents` | До скидок |
| `discountCents` | Сумма скидки |
| `totalPriceCents` | После скидок |
| `appliedPromocode` | Текущий код (если есть) |
| `comboDiscountApplied` | Сработала ли combo-правило |
| `expiresAt` | Когда истечёт TTL (RFC3339) |

### 2.1. Получить корзину

```http
GET /v1/users/{user_id}/cart
```

Пустая корзина → `items: []`, суммы `0`.

### 2.2. Добавить товар

```http
POST /v1/users/{user_id}/cart/items
```

```json
{ "product_id": "550e8400-e29b-41d4-a716-446655440001", "quantity": 1 }
```

Ожидание: повторное добавление того же товара **увеличивает** quantity; лимит 1–99; нельзя превысить `stockQuantity`; неизвестный товар → **404**.

```bash
curl -X POST "http://localhost:8080/v1/users/${USER_ID}/cart/items" \
  -H "Authorization: Bearer ${TOKEN}" \
  -H 'Content-Type: application/json' \
  -d '{"product_id":"550e8400-e29b-41d4-a716-446655440001","quantity":1}'
```

### 2.3. Удалить товар

```http
DELETE /v1/users/{user_id}/cart/items/{product_id}
```

Ожидание: товара нет в корзине → **404**.

### 2.4. Очистить корзину

```http
DELETE /v1/users/{user_id}/cart
```

Ожидание: пустые `items` и сброшенный промокод.

### 2.5. Применить промокод

```http
POST /v1/users/{user_id}/cart/promocode
```

```json
{ "code": "SAVE10" }
```

Неизвестный / неактивный / просроченный код → **404**.

### 2.6. Снять промокод

```http
DELETE /v1/users/{user_id}/cart/promocode
```

---

## 3. Заказы (JWT)

### Статусы

| JSON | Число | Как появляется |
|------|-------|----------------|
| `ORDER_STATUS_CREATED` | 1 | После `POST /v1/orders` |
| `ORDER_STATUS_PAID` | 2 | Ручной `UpdateOrderStatus` из `CREATED` |
| `ORDER_STATUS_SHIPPED` | 3 | Авто через **10 мин** после `PAID` (на `GET` заказа) |
| `ORDER_STATUS_CANCELLED` | 4 | Cancel / status из `CREATED` |
| `ORDER_STATUS_COMPLETED` | 5 | Авто через **10 мин** после `SHIPPED` |

Ручные переходы только: `CREATED` → `PAID`, `CREATED` → `CANCELLED`.  
`fromStatus` должен совпадать с текущим → иначе **FailedPrecondition**.  
Чужой заказ → **403**.

### 3.1. Создать заказ

```http
POST /v1/orders
```

```json
{ "user_id": "<uuid>" }
```

После успеха корзина очищается. Ожидание: пустая корзина → ошибка предусловия (не заказ на 0).

### 3.2. Получить заказ

```http
GET /v1/orders/{order_id}
```

При `GET` сервер может продвинуть `PAID`→`SHIPPED` / `SHIPPED`→`COMPLETED` по таймерам.

### 3.3. Отменить заказ

```http
POST /v1/orders/{order_id}/cancel
```

Только из `CREATED`.

### 3.4. Обновить статус

```http
POST /v1/orders/{order_id}/status
```

```json
{
  "fromStatus": "ORDER_STATUS_CREATED",
  "toStatus": "ORDER_STATUS_PAID"
}
```

### 3.5. Удалить заказ

```http
DELETE /v1/orders/{order_id}
```

Жёсткое удаление (строка + `order_items`). Любой статус. Владелец или admin. Отличается от soft-cancel (`POST .../cancel`).

---

## 4. Admin-промокоды (JWT admin)

```http
GET    /v1/admin/promocodes
POST   /v1/admin/promocodes
PATCH  /v1/admin/promocodes/{code}
DELETE /v1/admin/promocodes/{code}
```

Создание:

```json
{
  "code": "SPRING20",
  "discountType": "DISCOUNT_TYPE_PERCENT",
  "discountValue": "20",
  "active": true
}
```

Новые коды создаются **active**; деактивация — через `PATCH`.  
`discountType`: `DISCOUNT_TYPE_PERCENT` или `DISCOUNT_TYPE_FIXED_CENTS`.  
Опционально `expiresAt` (RFC3339).

```bash
ADMIN_TOKEN=$(curl -s http://localhost:8080/v1/users/login \
  -H 'Content-Type: application/json' \
  -d '{"email":"admin@store.local","password":"admin123"}' | jq -r .accessToken)

curl http://localhost:8080/v1/admin/promocodes \
  -H "Authorization: Bearer ${ADMIN_TOKEN}"
```

---

## Пример полного сценария (REST)

```bash
# Регистрация / логин
curl -s http://localhost:8080/v1/users/register \
  -H 'Content-Type: application/json' \
  -d '{"email":"qa@example.com","password":"secret123","name":"QA"}'

TOKEN=$(curl -s http://localhost:8080/v1/users/login \
  -H 'Content-Type: application/json' \
  -d '{"email":"qa@example.com","password":"secret123"}' | jq -r .accessToken)
USER_ID=$(curl -s http://localhost:8080/v1/users/login \
  -H 'Content-Type: application/json' \
  -d '{"email":"qa@example.com","password":"secret123"}' | jq -r .user.id)

# Каталог
curl http://localhost:8080/v1/products

# Корзина + промо
curl -X POST "http://localhost:8080/v1/users/${USER_ID}/cart/items" \
  -H "Authorization: Bearer ${TOKEN}" -H 'Content-Type: application/json' \
  -d '{"product_id":"550e8400-e29b-41d4-a716-446655440001","quantity":1}'

curl -X POST "http://localhost:8080/v1/users/${USER_ID}/cart/promocode" \
  -H "Authorization: Bearer ${TOKEN}" -H 'Content-Type: application/json' \
  -d '{"code":"SAVE10"}'

curl "http://localhost:8080/v1/users/${USER_ID}/cart" \
  -H "Authorization: Bearer ${TOKEN}"

# Заказ + оплата
ORDER_ID=$(curl -s -X POST http://localhost:8080/v1/orders \
  -H "Authorization: Bearer ${TOKEN}" -H 'Content-Type: application/json' \
  -d "{\"user_id\":\"${USER_ID}\"}" | jq -r .order.id)

curl -X POST "http://localhost:8080/v1/orders/${ORDER_ID}/status" \
  -H "Authorization: Bearer ${TOKEN}" -H 'Content-Type: application/json' \
  -d '{"fromStatus":"ORDER_STATUS_CREATED","toStatus":"ORDER_STATUS_PAID"}'

curl "http://localhost:8080/v1/orders/${ORDER_ID}" \
  -H "Authorization: Bearer ${TOKEN}"
```

---

# gRPC API

Адрес: `localhost:50051`  
Reflection включён. Контракты: `proto/*.proto`.

| Сервис | Пакет | Описание |
|--------|-------|----------|
| `UserService` | `store.user.v1` | Пользователи |
| `CatalogService` | `store.catalog.v1` | Каталог |
| `CartService` | `store.cart.v1` | Корзина |
| `OrderService` | `store.order.v1` | Заказы |
| `PromoService` | `store.promo.v1` | Admin-промокоды |

Для Cart / Order / Promo передавайте metadata: `authorization: Bearer <token>`.

### Product

| Поле | Тип | Описание |
|------|-----|----------|
| `id` | string | ID |
| `name` | string | Название |
| `description` | string | Описание |
| `price_cents` | int64 | Цена |
| `stock_quantity` | int32 | Остаток |
| `brand` | string | Бренд |

### CartResponse

| Поле | Тип | Описание |
|------|-----|----------|
| `items` | repeated CartItem | Позиции |
| `total_price_cents` | int64 | Итого |
| `subtotal_cents` | int64 | До скидок |
| `discount_cents` | int64 | Скидка |
| `applied_promocode` | string | Код |
| `combo_discount_applied` | bool | Combo |
| `expires_at` | string | TTL (RFC3339) |

### OrderStatus

| Имя | Значение |
|-----|----------|
| `ORDER_STATUS_CREATED` | 1 |
| `ORDER_STATUS_PAID` | 2 |
| `ORDER_STATUS_SHIPPED` | 3 |
| `ORDER_STATUS_CANCELLED` | 4 |
| `ORDER_STATUS_COMPLETED` | 5 |

### CartService (дополнительно)

- `ApplyPromocode` — `user_id`, `code`
- `ClearPromocode` — `user_id`

### OrderService (дополнительно)

- `CancelOrder` — `order_id`
- `UpdateOrderStatus` — `order_id`, `from_status`, `to_status`
- `DeleteOrder` — `order_id` (hard delete, любой статус)

### PromoService

- `CreatePromocode` / `ListPromocodes` / `UpdatePromocode` / `DeletePromocode`

---

## Соответствие REST ↔ gRPC

| REST | gRPC |
|------|------|
| `POST /v1/users/register` | `UserService.Register` |
| `POST /v1/users/login` | `UserService.Login` |
| `GET /v1/users/{user_id}` | `UserService.GetUser` |
| `DELETE /v1/users/{user_id}` | `UserService.DeleteUser` |
| `GET /v1/products` | `CatalogService.ListProducts` |
| `GET /v1/products/{product_id}` | `CatalogService.GetProduct` |
| `GET /v1/users/{user_id}/cart` | `CartService.GetCart` |
| `POST /v1/users/{user_id}/cart/items` | `CartService.AddItem` |
| `DELETE /v1/users/{user_id}/cart/items/{product_id}` | `CartService.RemoveItem` |
| `DELETE /v1/users/{user_id}/cart` | `CartService.ClearCart` |
| `POST /v1/users/{user_id}/cart/promocode` | `CartService.ApplyPromocode` |
| `DELETE /v1/users/{user_id}/cart/promocode` | `CartService.ClearPromocode` |
| `POST /v1/orders` | `OrderService.CreateOrder` |
| `GET /v1/orders/{order_id}` | `OrderService.GetOrder` |
| `POST /v1/orders/{order_id}/cancel` | `OrderService.CancelOrder` |
| `POST /v1/orders/{order_id}/status` | `OrderService.UpdateOrderStatus` |
| `DELETE /v1/orders/{order_id}` | `OrderService.DeleteOrder` |
| `GET /v1/admin/promocodes` | `PromoService.ListPromocodes` |
| `POST /v1/admin/promocodes` | `PromoService.CreatePromocode` |
| `PATCH /v1/admin/promocodes/{code}` | `PromoService.UpdatePromocode` |
| `DELETE /v1/admin/promocodes/{code}` | `PromoService.DeletePromocode` |

---

## Полезные команды grpcurl

```bash
grpcurl -plaintext localhost:50051 list

grpcurl -plaintext localhost:50051 describe store.cart.v1.CartService

grpcurl -plaintext \
  -H 'authorization: Bearer <token>' \
  -d '{"user_id":"<uuid>"}' \
  localhost:50051 store.cart.v1.CartService/GetCart
```
