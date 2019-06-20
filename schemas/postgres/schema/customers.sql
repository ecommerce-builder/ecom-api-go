CREATE TABLE IF NOT EXISTS customers (
  id             SERIAL PRIMARY KEY,
  uuid           UUID DEFAULT uuid_generate_v4() UNIQUE,
  uid            VARCHAR(64) NOT NULL UNIQUE,
  role           VARCHAR(64) NOT NULL,
  email          VARCHAR(512) NOT NULL UNIQUE,
  firstname      VARCHAR(255) NOT NULL,
  lastname       VARCHAR(255) NOT NULL,
  created        TIMESTAMP NOT NULL DEFAULT NOW(),
  modified       TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_customers_role_asc ON customers (role ASC);
CREATE INDEX IF NOT EXISTS idx_customers_created_desc ON customers (created DESC);
CREATE INDEX IF NOT EXISTS idx_customers_modified ON customers (modified DESC);
