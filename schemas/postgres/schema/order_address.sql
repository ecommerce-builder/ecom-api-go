CREATE TABLE IF NOT EXISTS order_address (
  id              SERIAL PRIMARY KEY,
  uuid            UUID DEFAULT uuid_generate_v4() UNIQUE,
  typ             address_t NOT NULL,
  contact_name    VARCHAR(1024) NOT NULL,
  addr1           VARCHAR(1024) NOT NULL,
  addr2           VARCHAR(1024),
  city            VARCHAR(512) NOT NULL,
  county          VARCHAR(512),
  postcode        VARCHAR(64) NOT NULL,
  country_code    CHAR(2) NOT NULL,
  created         TIMESTAMP NOT NULL DEFAULT NOW(),
  modified        TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS order_address_created_idx  ON order_address (created DESC);
CREATE INDEX IF NOT EXISTS order_address_modified_idx ON order_address (modified DESC);
