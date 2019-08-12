CREATE TABLE IF NOT EXISTS product (
  id            SERIAL PRIMARY KEY,
  uuid          UUID DEFAULT uuid_generate_v4() NOT NULL UNIQUE,
  sku           VARCHAR(64) NOT NULL UNIQUE,
  ean           VARCHAR(64) NOT NULL,
  path          VARCHAR(512) NOT NULL UNIQUE,
  name          VARCHAR(1024) NOT NULL,
  content       JSONB,
  created       TIMESTAMP NOT NULL DEFAULT NOW(),
  modified      TIMESTAMP NOT NULL DEFAULT NOW(),
  UNIQUE (id, sku),
  UNIQUE (sku, ean)
);

CREATE INDEX IF NOT EXISTS idx_product_created_desc ON product (created DESC);
CREATE INDEX IF NOT EXISTS idx_product_modified ON product (modified DESC);
