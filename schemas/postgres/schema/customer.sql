CREATE TABLE IF NOT EXISTS customer (
  id               SERIAL PRIMARY KEY,
  uuid             UUID DEFAULT uuid_generate_v4() UNIQUE,
  uid              VARCHAR(64) NOT NULL UNIQUE,
  pricing_tier_id  INTEGER NOT NULL DEFAULT 1,
  role             VARCHAR(64) NOT NULL,
  email            VARCHAR(512) NOT NULL UNIQUE,
  firstname        VARCHAR(255) NOT NULL,
  lastname         VARCHAR(255) NOT NULL,
  created          TIMESTAMP NOT NULL DEFAULT NOW(),
  modified         TIMESTAMP NOT NULL DEFAULT NOW(),
  FOREIGN KEY (pricing_tier_id) REFERENCES pricing_tier (id)
);

CREATE INDEX IF NOT EXISTS idx_customer_role_asc ON customer (role ASC);
CREATE INDEX IF NOT EXISTS idx_customer_created_desc ON customer (created DESC);
CREATE INDEX IF NOT EXISTS idx_customer_modified ON customer (modified DESC);
