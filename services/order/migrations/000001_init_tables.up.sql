CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TYPE item AS (
  item_id UUID,
  quantity INTEGER
);

CREATE TABLE IF NOT EXISTS orders(
  id UUID PRIMARY KEY,
  user_id UUID NOT NULL,      
  status VARCHAR(255) NOT NULL,
  currency VARCHAR(255) NOT NULL,
  total_price DECIMAL(10, 2) NOT NULL,
  payment_method VARCHAR(255) NOT NULL,
  delivery_method VARCHAR(255) NOT NULL,
  delivery_address VARCHAR(255) NOT NULL,
  delivery_date TIMESTAMP NOT NULL,
  items item[] NOT NULL,
  created_at TIMESTAMP NOT NULL,
  updated_at TIMESTAMP NOT NULL
);

CREATE TABLE IF NOT EXISTS coupons(
  id SERIAL PRIMARY KEY,
  code VARCHAR(255) NOT NULL UNIQUE,
  discount DECIMAL(10, 2) NOT NULL,
  valid_from TIMESTAMP NOT NULL,
  valid_to TIMESTAMP NOT NULL
);


INSERT INTO coupons (id, code, discount, valid_from, valid_to)
VALUES (1, 'TEST', 50.0, '2025-01-01 00:00:00', '2025-12-31 23:59:59');
