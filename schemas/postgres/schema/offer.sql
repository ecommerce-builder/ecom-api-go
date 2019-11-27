CREATE TABLE IF NOT EXISTS offer (
  id             SERIAL PRIMARY KEY,
  uuid           UUID UNIQUE NOT NULL DEFAULT uuid_generate_v4(),
  promo_rule_id  INTEGER NOT NULL UNIQUE,
  created        TIMESTAMP NOT NULL DEFAULT NOW(),
  modified       TIMESTAMP NOT NULL DEFAULT NOW(),
  FOREIGN KEY (promo_rule_id) REFERENCES promo_rule (id)
);
