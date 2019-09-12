CREATE TABLE shipping_tarrif (
  id             SERIAL PRIMARY KEY,
  uuid           UUID UNIQUE NOT NULL DEFAULT uuid_generate_v4(),
  country_code   CHAR(2) NOT NULL,
  shipping_code  VARCHAR(256) NOT NULL UNIQUE,
  name           VARCHAR(512),
  price          INTEGER NOT NULL CHECK (price >= 0),
  tax_code       VARCHAR(32) NULL DEFAULT NULL,
  UNIQUE (country_code, shipping_code),
  created        TIMESTAMP NOT NULL DEFAULT NOW(),
  modified       TIMESTAMP NOT NULL DEFAULT NOW()
);
