CREATE TABLE IF NOT EXISTS customers_devkeys (
  id           SERIAL PRIMARY KEY,
  uuid         UUID DEFAULT uuid_generate_v4() UNIQUE,
  key          VARCHAR(256),
  hash         VARCHAR(256),
  customer_id  INTEGER NOT NULL,
  created      TIMESTAMP NOT NULL DEFAULT NOW(),
  modified     TIMESTAMP NOT NULL DEFAULT NOW(),
  FOREIGN KEY (customer_id) REFERENCES customers(id),
  UNIQUE (customer_id, key)
);

CREATE INDEX IF NOT EXISTS idx_customers_devkeys_apikey_asc ON customers_devkeys (created DESC);
CREATE INDEX IF NOT EXISTS idx_customers_devkeys_created_desc ON customers_devkeys (created DESC);
CREATE INDEX IF NOT EXISTS idx_customers_devkeys_modified ON customers_devkeys (modified DESC);

ALTER SEQUENCE customers_devkeys_id_seq RESTART WITH 900001;
