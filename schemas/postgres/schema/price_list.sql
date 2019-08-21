DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'price_list_strategy_t') THEN
        CREATE TYPE price_list_strategy_t AS ENUM ('simple', 'volume', 'tiered');
    END IF;
END$$;

CREATE TABLE IF NOT EXISTS price_list (
  id             SERIAL PRIMARY KEY,
  uuid           UUID DEFAULT uuid_generate_v4() NOT NULL UNIQUE,
  code           VARCHAR(64) NOT NULL UNIQUE,
  currency_code  CHAR(3) NOT NULL DEFAULT 'GBP',
  strategy       price_list_strategy_t NOT NULL DEFAULT 'simple',
  inc_tax        BOOL NOT NULL DEFAULT false,
  name           VARCHAR(256) NOT NULL,
  description    VARCHAR(256) NULL,
  created        TIMESTAMP NOT NULL DEFAULT NOW(),
  modified       TIMESTAMP NOT NULL DEFAULT NOW()
);

INSERT INTO price_list (code, name, description) VALUES ('default', 'Default price list', 'Default price list');
