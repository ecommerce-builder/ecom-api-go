CREATE TABLE IF NOT EXISTS product_pricing (
  id           SERIAL PRIMARY KEY,
  tier_ref     VARCHAR(64) NOT NULL,
  sku          VARCHAR(64) NOT NULL,
  unit_price   INTEGER NOT NULL CHECK (unit_price >= 0),
  created      TIMESTAMP NOT NULL DEFAULT NOW(),
  modified     TIMESTAMP NOT NULL DEFAULT NOW(),
  FOREIGN KEY (tier_ref) REFERENCES pricing_tiers (tier_ref),
  FOREIGN KEY (sku) REFERENCES products (sku),
  UNIQUE (tier_ref, sku)
);
