# EducationProjectQA

Project for training Auto QA skills with API (REST + gRPC), Database, UI, Queues.

## Store API Simulator

gRPC store backend with a JSON HTTP gateway. Catalog, cart, and order services backed by **PostgreSQL**.

## Features

- **Catalog** — list and fetch products
- **Cart** — add/remove items, get cart, clear cart
- **Orders** — create an order from a cart, fetch order by ID
- **Dual transport** — native gRPC (`:50051`) and REST/JSON via grpc-gateway (`:8080`)
- **PostgreSQL** — persistent storage via `pgx` + `database/sql`

## Requirements

- Go **1.26+**
- Docker & Docker Compose (Postgres runs in Docker — no local Postgres install needed)
- For codegen: `protoc`, `protoc-gen-go`, `protoc-gen-go-grpc`, `protoc-gen-grpc-gateway`

## Quick start

### Docker (API + Postgres)

```bash
docker compose up --build
```

- gRPC: `localhost:50051`
- HTTP: `localhost:8080`
- Postgres: `localhost:5432` (user/password/db: `store` / `store` / `store`)

On first start Docker pulls the Postgres image, creates the database, and applies SQL from `migrations/` (schema + seed catalog).

### Local (API on host, Postgres in Docker)

```bash
docker compose up -d postgres
DATABASE_URL='postgres://store:store@localhost:5432/store?sslmode=disable' go run ./cmd/server
```

Default `DATABASE_URL` if unset: `postgres://store:store@localhost:5432/store?sslmode=disable`.

## Project layout

```
cmd/server/          # entrypoint (gRPC + HTTP gateway)
proto/               # Catalog, Cart, Order contracts
gen/                 # generated Go / gRPC / gateway code
migrations/          # Postgres schema + seed (applied on first DB init)
internal/
  handler/           # gRPC handlers
  service/           # business logic
  repository/        # models + postgres repos
third_party/         # googleapis for HTTP annotations
```

## Seed catalog

50 products: **15 Apple**, **15 Samsung**, **10 NVIDIA**, **10 AMD**.

| Brand   | ID range (suffix) | Examples |
|---------|-------------------|----------|
| Apple   | `...001`–`...015` | iPhone 15, MacBook Pro 14 M3, AirPods Pro 2 |
| Samsung | `...016`–`...030` | Galaxy S24 Ultra, Odyssey G9, 990 PRO |
| NVIDIA  | `...031`–`...040` | RTX 4090, RTX 4070 SUPER, Jetson Orin Nano |
| AMD     | `...041`–`...050` | Ryzen 9 7950X, RX 7900 XTX, EPYC 9654 |

Full UUID prefix: `550e8400-e29b-41d4-a716-44665544` + `0001`…`0050`. See [`migrations/002_seed.sql`](migrations/002_seed.sql).

## HTTP API

Base URL: `http://localhost:8080`

### Catalog

```http
GET /v1/products
GET /v1/products/{product_id}
```

### Cart

```http
GET    /v1/users/{user_id}/cart
POST   /v1/users/{user_id}/cart/items
DELETE /v1/users/{user_id}/cart/items/{product_id}
DELETE /v1/users/{user_id}/cart
```

Add item body:

```json
{
  "product_id": "550e8400-e29b-41d4-a716-446655440001",
  "quantity": 1
}
```

### Orders

```http
POST /v1/orders
GET  /v1/orders/{order_id}
```

Create order body:

```json
{
  "user_id": "user-1"
}
```

### Example flow

```bash
# List products
curl http://localhost:8080/v1/products

# Add to cart
curl -X POST http://localhost:8080/v1/users/user-1/cart/items \
  -H 'Content-Type: application/json' \
  -d '{"product_id":"550e8400-e29b-41d4-a716-446655440001","quantity":1}'

# Create order from cart
curl -X POST http://localhost:8080/v1/orders \
  -H 'Content-Type: application/json' \
  -d '{"user_id":"user-1"}'
```

## gRPC

Services (with reflection enabled):

- `store.catalog.v1.CatalogService`
- `store.cart.v1.CartService`
- `store.order.v1.OrderService`

Connect to `localhost:50051`. Contracts live in `proto/`.

## Code generation

Regenerate stubs after changing `.proto` files:

```bash
make generate
```

Install plugins if needed:

```bash
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway@latest
```

## Ports

| Port  | Protocol              |
|-------|-----------------------|
| 50051 | gRPC                  |
| 8080  | HTTP JSON (gateway)   |
| 5432  | PostgreSQL            |
