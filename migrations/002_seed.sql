-- Fixed UUIDs so API examples and docs stay stable across restarts.
-- 15 Apple, 15 Samsung, 10 NVIDIA, 10 AMD
INSERT INTO products (id, name, description, price_cents, stock_quantity)
VALUES
    -- Apple (15)
    ('550e8400-e29b-41d4-a716-446655440001', 'iPhone 15', 'Apple smartphone with A16 Bionic', 79900, 40),
    ('550e8400-e29b-41d4-a716-446655440002', 'iPhone 15 Pro', 'Apple flagship with titanium frame', 99900, 25),
    ('550e8400-e29b-41d4-a716-446655440003', 'iPhone 15 Pro Max', 'Largest iPhone 15 Pro display', 119900, 18),
    ('550e8400-e29b-41d4-a716-446655440004', 'MacBook Air 13 M3', 'Thin laptop with M3 chip', 109900, 20),
    ('550e8400-e29b-41d4-a716-446655440005', 'MacBook Pro 14 M3', 'Pro laptop for creators', 159900, 12),
    ('550e8400-e29b-41d4-a716-446655440006', 'MacBook Pro 16 M3 Max', 'High-end MacBook Pro', 349900, 6),
    ('550e8400-e29b-41d4-a716-446655440007', 'iPad Pro 11', 'OLED tablet with M4', 99900, 15),
    ('550e8400-e29b-41d4-a716-446655440008', 'iPad Air', 'Everyday Apple tablet', 59900, 30),
    ('550e8400-e29b-41d4-a716-446655440009', 'iPad mini', 'Compact Apple tablet', 49900, 22),
    ('550e8400-e29b-41d4-a716-446655440010', 'AirPods Pro 2', 'Noise-cancelling earbuds', 24900, 50),
    ('550e8400-e29b-41d4-a716-446655440011', 'AirPods Max', 'Over-ear Apple headphones', 54900, 14),
    ('550e8400-e29b-41d4-a716-446655440012', 'Apple Watch Ultra 2', 'Rugged GPS smartwatch', 79900, 16),
    ('550e8400-e29b-41d4-a716-446655440013', 'Apple Watch SE', 'Affordable fitness watch', 24900, 35),
    ('550e8400-e29b-41d4-a716-446655440014', 'Studio Display', '27-inch 5K Retina display', 159900, 8),
    ('550e8400-e29b-41d4-a716-446655440015', 'Magic Keyboard', 'Wireless keyboard with Touch ID', 14900, 40),

    -- Samsung (15)
    ('550e8400-e29b-41d4-a716-446655440016', 'Galaxy S24', 'Samsung flagship smartphone', 79900, 35),
    ('550e8400-e29b-41d4-a716-446655440017', 'Galaxy S24+', 'Larger Galaxy S24 display', 99900, 22),
    ('550e8400-e29b-41d4-a716-446655440018', 'Galaxy S24 Ultra', 'S Pen flagship phone', 129900, 15),
    ('550e8400-e29b-41d4-a716-446655440019', 'Galaxy Z Fold6', 'Foldable phone with large inner screen', 179900, 10),
    ('550e8400-e29b-41d4-a716-446655440020', 'Galaxy Z Flip6', 'Compact flip smartphone', 109900, 12),
    ('550e8400-e29b-41d4-a716-446655440021', 'Galaxy Tab S9', 'Android tablet for productivity', 79900, 18),
    ('550e8400-e29b-41d4-a716-446655440022', 'Galaxy Tab S9 Ultra', '14.6-inch Samsung tablet', 119900, 8),
    ('550e8400-e29b-41d4-a716-446655440023', 'Galaxy Watch7', 'Samsung smartwatch with Wear OS', 29900, 28),
    ('550e8400-e29b-41d4-a716-446655440024', 'Galaxy Watch Ultra', 'Rugged Samsung smartwatch', 64900, 14),
    ('550e8400-e29b-41d4-a716-446655440025', 'Galaxy Buds3 Pro', 'Samsung noise-cancelling earbuds', 24900, 40),
    ('550e8400-e29b-41d4-a716-446655440026', 'Galaxy Book4 Pro', 'Thin Windows laptop from Samsung', 149900, 11),
    ('550e8400-e29b-41d4-a716-446655440027', 'Odyssey G9', '49-inch ultrawide gaming monitor', 129900, 7),
    ('550e8400-e29b-41d4-a716-446655440028', 'Samsung 990 PRO 2TB', 'NVMe SSD for high-speed storage', 19900, 45),
    ('550e8400-e29b-41d4-a716-446655440029', 'The Frame 55"', 'Lifestyle 4K smart TV', 99900, 9),
    ('550e8400-e29b-41d4-a716-446655440030', 'Samsung T7 Shield 2TB', 'Rugged portable SSD', 17900, 33),

    -- NVIDIA (10)
    ('550e8400-e29b-41d4-a716-446655440031', 'GeForce RTX 4090', 'Flagship Ada Lovelace GPU', 159900, 5),
    ('550e8400-e29b-41d4-a716-446655440032', 'GeForce RTX 4080 SUPER', 'High-end 1440p/4K gaming GPU', 99900, 10),
    ('550e8400-e29b-41d4-a716-446655440033', 'GeForce RTX 4070 Ti SUPER', 'Strong 1440p graphics card', 79900, 14),
    ('550e8400-e29b-41d4-a716-446655440034', 'GeForce RTX 4070 SUPER', 'Efficient mid-high gaming GPU', 59900, 20),
    ('550e8400-e29b-41d4-a716-446655440035', 'GeForce RTX 4060 Ti', 'Popular 1080p/1440p GPU', 39900, 28),
    ('550e8400-e29b-41d4-a716-446655440036', 'GeForce RTX 4060', 'Entry Ada Lovelace gaming GPU', 29900, 35),
    ('550e8400-e29b-41d4-a716-446655440037', 'NVIDIA Jetson Orin Nano', 'Edge AI developer kit', 49900, 16),
    ('550e8400-e29b-41d4-a716-446655440038', 'NVIDIA Shield TV Pro', 'Android TV streaming box', 19900, 24),
    ('550e8400-e29b-41d4-a716-446655440039', 'DGX Spark', 'Compact AI workstation', 299900, 3),
    ('550e8400-e29b-41d4-a716-446655440040', 'NVIDIA ConnectX-7 NIC', 'High-speed networking adapter', 89900, 8),

    -- AMD (10)
    ('550e8400-e29b-41d4-a716-446655440041', 'Ryzen 9 7950X', '16-core desktop CPU', 54900, 18),
    ('550e8400-e29b-41d4-a716-446655440042', 'Ryzen 7 7800X3D', 'Gaming CPU with 3D V-Cache', 39900, 22),
    ('550e8400-e29b-41d4-a716-446655440043', 'Ryzen 5 7600', '6-core mainstream CPU', 19900, 40),
    ('550e8400-e29b-41d4-a716-446655440044', 'Ryzen AI 9 HX 370', 'AI laptop APU', 69900, 15),
    ('550e8400-e29b-41d4-a716-446655440045', 'Radeon RX 7900 XTX', 'Flagship RDNA 3 graphics card', 89900, 9),
    ('550e8400-e29b-41d4-a716-446655440046', 'Radeon RX 7800 XT', '1440p gaming GPU', 49900, 16),
    ('550e8400-e29b-41d4-a716-446655440047', 'Radeon RX 7600', '1080p gaming GPU', 26900, 30),
    ('550e8400-e29b-41d4-a716-446655440048', 'Ryzen Threadripper 7980X', '64-core HEDT CPU', 499900, 2),
    ('550e8400-e29b-41d4-a716-446655440049', 'AMD EPYC 9654', 'Server CPU for data centers', 899900, 4),
    ('550e8400-e29b-41d4-a716-446655440050', 'Ryzen 9 9950X', 'Zen 5 flagship desktop CPU', 64900, 12)
ON CONFLICT (id) DO NOTHING;
