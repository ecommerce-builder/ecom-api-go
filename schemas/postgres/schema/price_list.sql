CREATE TABLE IF NOT EXISTS price_list (
  id           SERIAL PRIMARY KEY,
  uuid         UUID DEFAULT uuid_generate_v4() NOT NULL UNIQUE,
  code         VARCHAR(64) NOT NULL UNIQUE,
  name         VARCHAR(256) NOT NULL UNIQUE,
  description  VARCHAR(256) NULL,
  created      TIMESTAMP NOT NULL DEFAULT NOW(),
  modified     TIMESTAMP NOT NULL DEFAULT NOW()
);

INSERT INTO price_list (code, name, description) VALUES ('default', 'Default price list', 'Default price list');
