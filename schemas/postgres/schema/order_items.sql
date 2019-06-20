CREATE TABLE IF NOT EXISTS order_items (
  id               SERIAL PRIMARY KEY,
  uuid             UUID DEFAULT uuid_generate_v4() UNIQUE,
  order_id         INTEGER NOT NULL,
  sku              VARCHAR(64) NOT NULL,
  qty              SMALLINT NOT NULL CHECK (qty >= 1 AND qty < 10000),
  unit_price       NUMERIC(8, 4) NOT NULL CHECK (unit_price >= 0.0000),
  vat              NUMERIC(8, 4) NOT NULL CHECK (vat >= 0.0000 AND vat <= 20.0000),
  created          TIMESTAMP NOT NULL DEFAULT NOW(),
  FOREIGN KEY (order_id) REFERENCES orders(id)
);
CREATE UNIQUE INDEX IF NOT EXISTS order_items_idx ON order_items (order_id, sku);
