CREATE TABLE IF NOT EXISTS products (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name            TEXT NOT NULL,
    description     TEXT NOT NULL DEFAULT '',
    price_cents     BIGINT NOT NULL,
    stock_quantity  INTEGER NOT NULL DEFAULT 0
);

CREATE TABLE IF NOT EXISTS cart_items (
    user_id     TEXT NOT NULL,
    product_id  UUID NOT NULL REFERENCES products (id),
    quantity    INTEGER NOT NULL CHECK (quantity > 0),
    PRIMARY KEY (user_id, product_id)
);

CREATE TABLE IF NOT EXISTS orders (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id             TEXT NOT NULL,
    total_amount_cents  BIGINT NOT NULL,
    status              INTEGER NOT NULL DEFAULT 1,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS order_items (
    order_id     UUID NOT NULL REFERENCES orders (id) ON DELETE CASCADE,
    product_id   UUID NOT NULL,
    quantity     INTEGER NOT NULL CHECK (quantity > 0),
    price_cents  BIGINT NOT NULL,
    PRIMARY KEY (order_id, product_id)
);
