CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE IF NOT EXISTS payments (
  id uuid PRIMARY KEY,
  user_id uuid NOT NULL,
  order_id uuid NOT NULL,
  currency VARCHAR(3) NOT NULL,
  total_price DECIMAL(10, 2) NOT NULL,
  status VARCHAR(50) NOT NULL,
  payment_method VARCHAR(50) NOT NULL,
  description VARCHAR(255) NOT NULL,
  created_at TIMESTAMPTZ NOT NULL,
  updated_at TIMESTAMPTZ NOT NULL
);

-- #FIXME убрать генерацию отсюда в application?
CREATE TABLE IF NOT EXISTS outbox (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  topic VARCHAR(100) NOT NULL,
  event_type VARCHAR(100) NOT NULL,
  order_id UUID NOT NULL,
  -- payload JSONB NOT NULL,
  created_at TIMESTAMP NOT NULL,
  processed_at TIMESTAMP
);
