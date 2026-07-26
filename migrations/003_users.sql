CREATE TABLE IF NOT EXISTS users (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email         TEXT NOT NULL UNIQUE,
    password_hash TEXT NOT NULL,
    name          TEXT NOT NULL DEFAULT '',
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- cart / orders must reference a real user (UUID)
-- Requires empty or UUID-only user_id data. Wipe volume if old text ids exist:
--   docker compose down -v
ALTER TABLE cart_items
    ALTER COLUMN user_id TYPE UUID USING user_id::uuid;

ALTER TABLE orders
    ALTER COLUMN user_id TYPE UUID USING user_id::uuid;

ALTER TABLE cart_items
    ADD CONSTRAINT cart_items_user_id_fkey
    FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE;

ALTER TABLE orders
    ADD CONSTRAINT orders_user_id_fkey
    FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE RESTRICT;
