CREATE TABLE IF NOT EXISTS order_item (
  id               SERIAL PRIMARY KEY,
  uuid             UUID DEFAULT uuid_generate_v4() UNIQUE,
  order_id         INTEGER NOT NULL,
  path             VARCHAR(1024) NOT NULL,
  sku              VARCHAR(64) NOT NULL,
  name             VARCHAR(1024) NOT NULL,
  qty              SMALLINT NOT NULL CHECK (qty >= 1 AND qty < 10000),
  unit_price       INTEGER NOT NULL CHECK (unit_price >= 0),
  currency         CHAR(3) NOT NULL DEFAULT 'GBP',
  discount         INTEGER DEFAULT NULL CHECK (discount >= 0 AND discount <= 10000),
  tax_code         VARCHAR(32) NULL DEFAULT NULL,
  vat              INTEGER NOT NULL CHECK (vat >= 0),
  created          TIMESTAMP NOT NULL DEFAULT NOW(),
  FOREIGN KEY (order_id) REFERENCES "order" (id)
);
CREATE UNIQUE INDEX IF NOT EXISTS order_item_idx ON order_item (order_id, sku);
