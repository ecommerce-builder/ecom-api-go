CREATE TABLE IF NOT EXISTS product_pricing (
  id               SERIAL PRIMARY KEY,
  uuid             UUID DEFAULT uuid_generate_v4() NOT NULL UNIQUE,
  pricing_tier_id  INTEGER,
  product_id       INTEGER,
  unit_price       INTEGER NOT NULL CHECK (unit_price >= 0),
  created          TIMESTAMP NOT NULL DEFAULT NOW(),
  modified         TIMESTAMP NOT NULL DEFAULT NOW(),
  FOREIGN KEY (pricing_tier_id) REFERENCES pricing_tier (id),
  FOREIGN KEY (product_id) REFERENCES product (id),
  UNIQUE (pricing_tier_id, product_id)
);
