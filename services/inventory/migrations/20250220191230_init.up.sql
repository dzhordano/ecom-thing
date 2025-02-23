CREATE TABLE IF NOT EXISTS items(
  product_id VARCHAR(255) NOT NULL,
  available_quantity BIGINT NOT NULL,
  reserved_quantity BIGINT NOT NULL,
  PRIMARY KEY (product_id)
);