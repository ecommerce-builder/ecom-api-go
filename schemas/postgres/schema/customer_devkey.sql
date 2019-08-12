CREATE TABLE IF NOT EXISTS customer_devkey (
  id           SERIAL PRIMARY KEY,
  uuid         UUID DEFAULT uuid_generate_v4() UNIQUE,
  customer_id  INTEGER NOT NULL,
  key          VARCHAR(256),
  hash         VARCHAR(256),
  created      TIMESTAMP NOT NULL DEFAULT NOW(),
  modified     TIMESTAMP NOT NULL DEFAULT NOW(),
  FOREIGN KEY (customer_id) REFERENCES customer (id),
  UNIQUE (customer_id, key)
);

CREATE INDEX IF NOT EXISTS idx_customer_devkey_apikey_asc ON customer_devkey (created DESC);
CREATE INDEX IF NOT EXISTS idx_customer_devkey_created_desc ON customer_devkey (created DESC);
CREATE INDEX IF NOT EXISTS idx_customer_devkey_modified ON customer_devkey (modified DESC);

ALTER SEQUENCE customer_devkey_id_seq RESTART WITH 900001;
