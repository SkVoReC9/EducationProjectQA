# Store API Simulator

gRPC store backend with a JSON HTTP gateway. In-memory catalog, cart, and order services for local development and API testing.

## Features

- **Catalog** — list and fetch products
- **Cart** — add/remove items, get cart, clear cart
- **Orders** — create an order from a cart, fetch order by ID
- **Dual transport** — native gRPC (`:50051`) and REST/JSON via grpc-gateway (`:8080`)
- **In-memory storage** — no database required; data resets on restart

## Requirements

- Go **1.26+**
- Docker & Docker Compose (optional)
- For codegen: `protoc`, `protoc-gen-go`, `protoc-gen-go-grpc`, `protoc-gen-grpc-gateway`

## Quick start

### Docker

```bash
docker compose up --build
```

- gRPC: `localhost:50051`
- HTTP: `localhost:8080`

### Local

```bash
go run ./cmd/server
```

## Project layout

```
cmd/server/          # entrypoint (gRPC + HTTP gateway)
proto/               # Catalog, Cart, Order contracts
gen/                 # generated Go / gRPC / gateway code
internal/
  handler/           # gRPC handlers
  service/           # business logic
  repository/        # models + in-memory repos
third_party/         # googleapis for HTTP annotations
```

## Seed catalog

| ID      | Name         | Price (cents) | Stock |
|---------|--------------|---------------|-------|
| prod-1  | iPhone 15    | 99900         | 10    |
| prod-2  | MacBook Pro  | 199900        | 5     |

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
  "product_id": "prod-1",
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
  -d '{"product_id":"prod-1","quantity":1}'

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
