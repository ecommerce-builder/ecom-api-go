CREATE TABLE IF NOT EXISTS cart_items (
  id          SERIAL PRIMARY KEY,
  uuid        UUID NOT NULL DEFAULT uuid_generate_v4(),
  cart_id     INTEGER,
  sku         VARCHAR(64) NOT NULL,
  qty         SMALLINT NOT NULL CHECK (qty >= 1 AND qty < 10000),
  unit_price  INTEGER NOT NULL CHECK (unit_price >= 0),
  created     TIMESTAMP NOT NULL DEFAULT NOW(),
  modified    TIMESTAMP NOT NULL DEFAULT NOW(),
  UNIQUE (cart_id, sku),
  FOREIGN KEY (cart_id) REFERENCES carts (id),
  FOREIGN KEY (sku) REFERENCES products (sku)
);
