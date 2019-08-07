CREATE TABLE IF NOT EXISTS offers (
  id             SERIAL PRIMARY KEY,
  uuid           UUID UNIQUE NOT NULL DEFAULT uuid_generate_v4(),
  promo_rule_id  INTEGER NOT NULL,
  created        TIMESTAMP NOT NULL DEFAULT NOW(),
  modified       TIMESTAMP NOT NULL DEFAULT NOW(),
  FOREIGN KEY (promo_rule_id) REFERENCES promo_rules (id)
);
