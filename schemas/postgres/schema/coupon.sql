CREATE TABLE IF NOT EXISTS coupon (
  id             SERIAL PRIMARY KEY,
  uuid           UUID UNIQUE NOT NULL DEFAULT uuid_generate_v4(),
  coupon_code    VARCHAR(64) NOT NULL UNIQUE,
  promo_rule_id  INTEGER NOT NULL,
  void           BOOLEAN NOT NULL DEFAULT 'f',
  reusable       BOOLEAN NOT NULL DEFAULT 'f',
  spend_count    INTEGER NOT NULL DEFAULT 0 CHECK (spend_count >= 0),
  created        TIMESTAMP NOT NULL DEFAULT NOW(),
  modified       TIMESTAMP NOT NULL DEFAULT NOW(),
  FOREIGN KEY (promo_rule_id) REFERENCES promo_rule (id)
);
