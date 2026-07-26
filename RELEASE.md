# Release notes — Store API Simulator

Changelog for mentors and QA mentees. Newest release first.  
After pulling a release that changes DB schema, recreate the volume:

```bash
docker compose down -v && docker compose up --build
```

API reference: [`Docs.MD`](Docs.MD) · Project overview: [`README.md`](README.md)

---

## Release 0.4.0 — 2026-07-26

**Theme:** Cart pricing, order lifecycle, admin promocodes

### Added

- **Promocodes (user):** `POST/DELETE /v1/users/{user_id}/cart/promocode`
- **Cart pricing response:** `subtotalCents`, `discountCents`, `appliedPromocode`, `comboDiscountApplied`, `expiresAt`
- **Combo rule:** NVIDIA + iPhone → 10% (see Docs.MD for expected rules)
- **Cart TTL:** 30 minutes inactivity
- **Order status:** `CancelOrder`, `UpdateOrderStatus` (`CREATED` → `PAID` / `CANCELLED`)
- **Auto progression:** `PAID` → `SHIPPED` (10m) → `COMPLETED` (10m), evaluated on `GET` order
- **Status** `ORDER_STATUS_COMPLETED`
- **Admin promocodes:** CRUD under `/v1/admin/promocodes` (`PromoService`)
- **Product `brand`:** `apple` / `samsung` / `nvidia` / `amd`
- **User `role`:** `user` \| `admin` in JWT; seed admin `admin@store.local` / `admin123`
- Seed promocodes: `SAVE10`, `FLAT500`, `WELCOME`

### Changed

- Cart & Order remain JWT-protected; admin routes require `role=admin`
- Order `GET` may advance status by timers
- [`Docs.MD`](Docs.MD) updated for mentees (auth, pricing, statuses, admin)

### Migrations

- `migrations/004_cart_pricing.sql` — `products.brand`, `carts`, `promocodes`, `users.role`, `orders.updated_at`, admin + promo seeds

### Breaking / QA impact

- Fresh DB volume required if upgrading from 0.3.x (`docker compose down -v`)
- Catalog JSON now includes `brand`
- Cart JSON has extra pricing fields
- Order status enum includes `COMPLETED`
- New endpoints to cover in regression suites (promo, cancel, status, admin)

---

## Release 0.3.0 — ~2026-07

**Theme:** Users & JWT auth

### Added

- `UserService`: register, login, get user
- JWT (`Authorization: Bearer`) for Cart and Order
- `user_id` must match token subject (403 otherwise)
- Migration `003_users.sql` — `users` table; cart/orders `user_id` as UUID FK

### Changed

- Cart/Order no longer accept anonymous string ids like `user-1`
- Flow: register/login → token → cart → order

### Breaking / QA impact

- All cart/order calls need a real registered user UUID + JWT
- Wipe DB if old text `user_id` values remain

---

## Release 0.2.0 — ~2026-06/07

**Theme:** PostgreSQL persistence

### Added

- Postgres via Docker Compose
- Migrations `001_init.sql`, `002_seed.sql` (50 products)
- Persistence for products, cart_items, orders / order_items

### Changed

- In-memory store replaced with SQL repositories
- Seed catalog stable UUIDs for docs and tests

### Breaking / QA impact

- Requires Docker Postgres (or local DSN)
- Data survives restarts (until volume wipe)

---

## Release 0.1.0 — init

**Theme:** First Store API simulator

### Added

- gRPC + HTTP gateway (`:50051` / `:8080`)
- Catalog: list / get product
- Cart: add / remove / get / clear
- Order: create from cart / get by id
- [`Docs.MD`](Docs.MD) API documentation for mentees

---

## How to add the next release

1. Bump section at the **top** (`Release x.y.z — YYYY-MM-DD`).
2. Fill **Added / Changed / Migrations / Breaking**.
3. Update [`Docs.MD`](Docs.MD) for mentees (expected behavior only).
4. Mentors: keep intentional-bug notes in gitignored `docs/BUGS.md` if needed.
