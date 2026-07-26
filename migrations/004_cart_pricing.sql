-- Brand on products
ALTER TABLE products
    ADD COLUMN IF NOT EXISTS brand TEXT NOT NULL DEFAULT '';

UPDATE products SET brand = 'apple'
WHERE id BETWEEN '550e8400-e29b-41d4-a716-446655440001'::uuid
          AND '550e8400-e29b-41d4-a716-446655440015'::uuid;

UPDATE products SET brand = 'samsung'
WHERE id BETWEEN '550e8400-e29b-41d4-a716-446655440016'::uuid
          AND '550e8400-e29b-41d4-a716-446655440030'::uuid;

UPDATE products SET brand = 'nvidia'
WHERE id BETWEEN '550e8400-e29b-41d4-a716-446655440031'::uuid
          AND '550e8400-e29b-41d4-a716-446655440040'::uuid;

UPDATE products SET brand = 'amd'
WHERE id BETWEEN '550e8400-e29b-41d4-a716-446655440041'::uuid
          AND '550e8400-e29b-41d4-a716-446655440050'::uuid;

-- Cart meta: promocode + TTL timestamp
CREATE TABLE IF NOT EXISTS carts (
    user_id    UUID PRIMARY KEY REFERENCES users (id) ON DELETE CASCADE,
    promocode  TEXT,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Promocodes (admin-managed)
CREATE TABLE IF NOT EXISTS promocodes (
    code            TEXT PRIMARY KEY,
    discount_type   TEXT NOT NULL CHECK (discount_type IN ('percent', 'fixed_cents')),
    discount_value  BIGINT NOT NULL CHECK (discount_value > 0),
    active          BOOLEAN NOT NULL DEFAULT TRUE,
    expires_at      TIMESTAMPTZ
);

INSERT INTO promocodes (code, discount_type, discount_value, active)
VALUES
    ('SAVE10', 'percent', 10, TRUE),
    ('FLAT500', 'fixed_cents', 500, TRUE),
    ('WELCOME', 'percent', 15, TRUE)
ON CONFLICT (code) DO NOTHING;

-- User roles
ALTER TABLE users
    ADD COLUMN IF NOT EXISTS role TEXT NOT NULL DEFAULT 'user';

-- Seed admin: admin@store.local / admin123
INSERT INTO users (id, email, password_hash, name, role)
VALUES (
    'aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee',
    'admin@store.local',
    '$2a$10$OHAbNHxnWEFXOokd7H/tEewQdz4wd5x9KUTKyMleKmenXvMbkkqLa',
    'Store Admin',
    'admin'
)
ON CONFLICT (email) DO UPDATE SET role = EXCLUDED.role;

-- Order status timestamps for auto PAID -> SHIPPED -> COMPLETED
ALTER TABLE orders
    ADD COLUMN IF NOT EXISTS updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW();

UPDATE orders SET updated_at = created_at;
