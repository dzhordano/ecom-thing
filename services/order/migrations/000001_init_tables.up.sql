CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TYPE item AS (
  item_id UUID,
  quantity INTEGER
);

CREATE TABLE IF NOT EXISTS orders(
  id UUID PRIMARY KEY,
  user_id UUID NOT NULL,      
  description VARCHAR(255) NOT NULL,
  status VARCHAR(255) NOT NULL,
  currency VARCHAR(255) NOT NULL,
  total_price DECIMAL(10, 2) NOT NULL,
  payment_method VARCHAR(255) NOT NULL,
  delivery_method VARCHAR(255) NOT NULL,
  delivery_address VARCHAR(255) NOT NULL,
  delivery_date TIMESTAMPTZ NOT NULL,
  items item[] NOT NULL,
  created_at TIMESTAMPTZ NOT NULL,
  updated_at TIMESTAMPTZ NOT NULL
);

CREATE TABLE IF NOT EXISTS coupons(
  id SERIAL PRIMARY KEY,
  code VARCHAR(255) NOT NULL UNIQUE,
  discount DECIMAL(10, 2) NOT NULL,
  valid_from TIMESTAMPTZ NOT NULL,
  valid_to TIMESTAMPTZ NOT NULL
);

-- FIXME убрать генерацию отсюда?
CREATE TABLE IF NOT EXISTS outbox (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  topic VARCHAR(100) NOT NULL,
  event_type VARCHAR(100) NOT NULL,
  payload JSONB NOT NULL,
  created_at TIMESTAMPTZ NOT NULL,
  processed_at TIMESTAMPTZ
);

-- test coupon
-- INSERT INTO coupons (id, code, discount, valid_from, valid_to)
-- VALUES (1, 'TEST', 50.0, '2025-01-01 00:00:00', '2025-12-31 23:59:59');

-- test order
-- INSERT INTO orders (id, user_id, description, status, currency, total_price, payment_method, delivery_method, delivery_address, delivery_date, items, created_at, updated_at)
-- VALUES ('00000000-0000-0000-0000-000000000001',
--  '00000000-0000-0000-0000-000000000002',
--  'Test Description For Test Order Bruh Test Test Test Test Test',
--  'pending', 'RUB', 100.00, 'card', 'standard', 'Address', '2077-01-01 00:00:00',
--  ARRAY[
--   ('00000000-0000-0000-0000-000000000003',1), 
--   ('00000000-0000-0000-0000-000000000004',2)]::item[],
--   '2025-01-01 00:00:00', '2025-01-01 00:00:00');
