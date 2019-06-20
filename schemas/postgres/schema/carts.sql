CREATE TABLE IF NOT EXISTS carts (
  id          SERIAL PRIMARY KEY,
  uuid        UUID NOT NULL,
  sku         VARCHAR(64) NOT NULL,
  qty         SMALLINT NOT NULL CHECK (qty >= 1 AND qty < 10000),
  unit_price  NUMERIC(8, 4) NOT NULL CHECK (unit_price >= 0.0000),
  created     TIMESTAMP NOT NULL DEFAULT NOW(),
  modified    TIMESTAMP NOT NULL DEFAULT NOW(),
  UNIQUE (uuid, sku),
  FOREIGN KEY (sku) REFERENCES products (sku)
);
