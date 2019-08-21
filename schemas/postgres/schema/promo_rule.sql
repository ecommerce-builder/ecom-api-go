DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'promo_rule_type_t') THEN
        CREATE TYPE promo_rule_type_t AS ENUM ('percentage', 'fixed');
    END IF;

    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'promo_rule_target_t') THEN
        CREATE TYPE promo_rule_target_t AS ENUM ('product', 'productset', 'category', 'total', 'shipping');
    END IF;
END$$;

CREATE TABLE IF NOT EXISTS promo_rule (
  id                 SERIAL PRIMARY KEY,
  uuid               UUID UNIQUE NOT NULL DEFAULT uuid_generate_v4(),
  product_id         INTEGER NULL DEFAULT NULL,
  product_set_id     INTEGER NULL DEFAULT NULL,
  category_id        INTEGER NULL DEFAULT NULL,
  shipping_tarrif_id INTEGER NULL DEFAULT NULL,
  name               VARCHAR(255) NOT NULL DEFAULT '',
  starts_at          TIMESTAMP NULL,
  ends_at            TIMESTAMP NULL,
  amount             INTEGER NOT NULL CHECK (amount >= 0),
  total_threshold    INTEGER CHECK (total_threshold >= 0),
  type               promo_rule_type_t NOT NULL,
  target             promo_rule_target_t NOT NULL,
  created            TIMESTAMP NOT NULL DEFAULT NOW(),
  modified           TIMESTAMP NOT NULL DEFAULT NOW(),
  FOREIGN KEY        (product_id) REFERENCES product (id),
  FOREIGN KEY        (product_set_id) REFERENCES product_set (id),
  FOREIGN KEY        (category_id) REFERENCES category (id),
  FOREIGN KEY        (shipping_tarrif_id) REFERENCES shipping_tarrif (id)
);
