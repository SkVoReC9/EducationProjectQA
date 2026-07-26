# EducationProjectQA

Project for training Auto QA skills with API (REST + gRPC), Database, UI, Queues.

**QA mentees:** API reference and expected behavior → [`Docs.MD`](Docs.MD).

## Store API Simulator

gRPC store backend with a JSON HTTP gateway. Catalog, cart, order, user, and admin promocode services backed by **PostgreSQL**.

## Features

- **Users** — register, login (JWT + role), get user by ID
- **Catalog** — list and fetch products with `brand` (public)
- **Cart** — add/remove items, get/clear cart, apply/clear promocode, 30m TTL (**JWT required**)
- **Orders** — create from cart, get, cancel, update status; auto `PAID`→`SHIPPED`→`COMPLETED` (**JWT required**)
- **Admin promocodes** — CRUD under `/v1/admin/promocodes` (**admin JWT**)
- **Dual transport** — native gRPC (`:50051`) and REST/JSON via grpc-gateway (`:8080`)
- **PostgreSQL** — persistent storage via `pgx` + `database/sql`

## Requirements

- Go **1.26+**
- Docker & Docker Compose (Postgres runs in Docker — no local Postgres install needed)
- For codegen: `protoc`, `protoc-gen-go`, `protoc-gen-go-grpc`, `protoc-gen-grpc-gateway`

## Quick start

### Docker (API + Postgres)

```bash
docker compose down -v
docker compose up --build
```

Use `-v` when schema migrations changed so init scripts re-run on a fresh volume.

- gRPC: `localhost:50051`
- HTTP: `localhost:8080`
- Postgres: `localhost:5432` (user/password/db: `store` / `store` / `store`)

### Local (API on host, Postgres in Docker)

```bash
docker compose up -d postgres
DATABASE_URL='postgres://store:store@localhost:5432/store?sslmode=disable' \
JWT_SECRET='dev-secret-change-me' \
go run ./cmd/server
```

Defaults if unset:

- `DATABASE_URL` → `postgres://store:store@localhost:5432/store?sslmode=disable`
- `JWT_SECRET` → `dev-secret-change-me`

## Project layout

```
cmd/server/          # entrypoint (gRPC + HTTP gateway)
proto/               # Catalog, Cart, Order, User, Promo contracts
gen/                 # generated Go / gRPC / gateway code
migrations/          # Postgres schema + seed (applied on first DB init)
internal/
  auth/              # JWT manager + gRPC interceptor
  handler/           # gRPC handlers
  service/           # business logic
  repository/        # models + postgres repos
third_party/         # googleapis for HTTP annotations
```

## Seed catalog

50 products: **15 Apple**, **15 Samsung**, **10 NVIDIA**, **10 AMD** (`brand` field: `apple` / `samsung` / `nvidia` / `amd`).

| Brand   | ID range (suffix) | Examples |
|---------|-------------------|----------|
| Apple   | `...001`–`...015` | iPhone 15, MacBook Pro 14 M3, AirPods Pro 2 |
| Samsung | `...016`–`...030` | Galaxy S24 Ultra, Odyssey G9, 990 PRO |
| NVIDIA  | `...031`–`...040` | RTX 4090, RTX 4070 SUPER, Jetson Orin Nano |
| AMD     | `...041`–`...050` | Ryzen 9 7950X, RX 7900 XTX, EPYC 9654 |

Full UUID prefix: `550e8400-e29b-41d4-a716-44665544` + `0001`…`0050`.

## Auth

- **Public:** Catalog, `POST /v1/users/register`, `POST /v1/users/login`, `GET /v1/users/{user_id}`
- **Protected:** Cart, Order — `Authorization: Bearer <access_token>`
- **Admin:** Promo CRUD — JWT with `role=admin`
- Access token TTL: **24h** (HS256, `JWT_SECRET`); claims include `role`
- Path/body `user_id` on cart/order must match JWT `sub` (unless admin where noted)
- Seeded admin: `admin@store.local` / `admin123`

## Pricing rules (cart)

- **Subtotal** — sum of catalog prices × qty
- **Combo discount** — if cart has ≥1 `brand=nvidia` **and** ≥1 product whose name contains `iPhone` → **10%** off subtotal
- **Promocode** — applied via API; case-insensitive codes; **does not stack** with combo (promo wins)
- **ClearCart** — clears items and promocode
- **Cart TTL** — **30 minutes** of inactivity; refreshed on add/remove/clear/apply/clear promocode (not on get)
- Seeded codes: `SAVE10` (10%), `FLAT500` (500¢), `WELCOME` (15%)

Cart response includes `subtotalCents`, `discountCents`, `totalPriceCents`, `appliedPromocode`, `comboDiscountApplied`, `expiresAt`.

## Order status

| Status | How |
|--------|-----|
| `CREATED` | `POST /v1/orders` |
| `PAID` | `POST /v1/orders/{id}/status` with `fromStatus=CREATED`, `toStatus=PAID` |
| `CANCELLED` | `POST /v1/orders/{id}/cancel` or status update `CREATED`→`CANCELLED` |
| `SHIPPED` | automatic **10 minutes** after becoming `PAID` (evaluated on `GET` order) |
| `COMPLETED` | automatic **10 minutes** after becoming `SHIPPED` |

Manual transitions other than `CREATED`→`PAID` / `CREATED`→`CANCELLED` are rejected (`FailedPrecondition`). `fromStatus` must match the current status.

## HTTP API

Base URL: `http://localhost:8080`

### Users

```http
POST /v1/users/register
POST /v1/users/login
GET  /v1/users/{user_id}
```

### Catalog

```http
GET /v1/products
GET /v1/products/{product_id}
```

### Cart (JWT)

```http
GET    /v1/users/{user_id}/cart
POST   /v1/users/{user_id}/cart/items
DELETE /v1/users/{user_id}/cart/items/{product_id}
DELETE /v1/users/{user_id}/cart
POST   /v1/users/{user_id}/cart/promocode
DELETE /v1/users/{user_id}/cart/promocode
```

Add item:

```json
{ "product_id": "550e8400-e29b-41d4-a716-446655440001", "quantity": 1 }
```

Apply promocode:

```json
{ "code": "SAVE10" }
```

### Orders (JWT)

```http
POST /v1/orders
GET  /v1/orders/{order_id}
POST /v1/orders/{order_id}/cancel
POST /v1/orders/{order_id}/status
```

Create:

```json
{ "user_id": "<uuid>" }
```

Update status:

```json
{ "fromStatus": "ORDER_STATUS_CREATED", "toStatus": "ORDER_STATUS_PAID" }
```

(grpc-gateway may also accept numeric enum values.)

### Admin promocodes (admin JWT)

```http
GET    /v1/admin/promocodes
POST   /v1/admin/promocodes
PATCH  /v1/admin/promocodes/{code}
DELETE /v1/admin/promocodes/{code}
```

Create body:

```json
{
  "code": "SPRING20",
  "discountType": "DISCOUNT_TYPE_PERCENT",
  "discountValue": "20",
  "active": true
}
```

### Example flow

```bash
# Admin login
ADMIN_TOKEN=$(curl -s http://localhost:8080/v1/users/login \
  -H 'Content-Type: application/json' \
  -d '{"email":"admin@store.local","password":"admin123"}' | jq -r .accessToken)

# Register user
curl -s http://localhost:8080/v1/users/register \
  -H 'Content-Type: application/json' \
  -d '{"email":"qa@example.com","password":"secret123","name":"QA"}'

TOKEN=$(curl -s http://localhost:8080/v1/users/login \
  -H 'Content-Type: application/json' \
  -d '{"email":"qa@example.com","password":"secret123"}' | jq -r .accessToken)
USER_ID=$(curl -s http://localhost:8080/v1/users/login \
  -H 'Content-Type: application/json' \
  -d '{"email":"qa@example.com","password":"secret123"}' | jq -r .user.id)

# Add iPhone + NVIDIA GPU, apply promo
curl -X POST "http://localhost:8080/v1/users/${USER_ID}/cart/items" \
  -H "Authorization: Bearer ${TOKEN}" \
  -H 'Content-Type: application/json' \
  -d '{"product_id":"550e8400-e29b-41d4-a716-446655440001","quantity":1}'

curl -X POST "http://localhost:8080/v1/users/${USER_ID}/cart/items" \
  -H "Authorization: Bearer ${TOKEN}" \
  -H 'Content-Type: application/json' \
  -d '{"product_id":"550e8400-e29b-41d4-a716-446655440031","quantity":1}'

curl -X POST "http://localhost:8080/v1/users/${USER_ID}/cart/promocode" \
  -H "Authorization: Bearer ${TOKEN}" \
  -H 'Content-Type: application/json' \
  -d '{"code":"SAVE10"}'

ORDER_ID=$(curl -s -X POST http://localhost:8080/v1/orders \
  -H "Authorization: Bearer ${TOKEN}" \
  -H 'Content-Type: application/json' \
  -d "{\"user_id\":\"${USER_ID}\"}" | jq -r .order.id)

curl -X POST "http://localhost:8080/v1/orders/${ORDER_ID}/status" \
  -H "Authorization: Bearer ${TOKEN}" \
  -H 'Content-Type: application/json' \
  -d '{"fromStatus":"ORDER_STATUS_CREATED","toStatus":"ORDER_STATUS_PAID"}'
```

> grpc-gateway JSON uses camelCase (`accessToken`, `userId`, `fromStatus`, …).

## gRPC

- `store.user.v1.UserService`
- `store.catalog.v1.CatalogService`
- `store.cart.v1.CartService`
- `store.order.v1.OrderService`
- `store.promo.v1.PromoService`

Connect to `localhost:50051`. Contracts live in `proto/`.

## Code generation

```bash
make generate
```

## Ports

| Port  | Protocol              |
|-------|-----------------------|
| 50051 | gRPC                  |
| 8080  | HTTP JSON (gateway)   |
| 5432  | PostgreSQL            |
