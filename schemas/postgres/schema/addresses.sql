DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'address_t') THEN
        CREATE TYPE address_t AS ENUM('shipping', 'billing');
    END IF;
END$$;

CREATE TABLE IF NOT EXISTS addresses (
  id              SERIAL PRIMARY KEY,
  uuid            UUID DEFAULT uuid_generate_v4() UNIQUE,
  customer_id     INTEGER NOT NULL,
  typ             address_t NOT NULL,
  contact_name    VARCHAR(1024) NOT NULL,
  addr1           VARCHAR(1024) NOT NULL,
  addr2           VARCHAR(1024),
  city            VARCHAR(512) NOT NULL,
  county          VARCHAR(512),
  postcode        VARCHAR(64) NOT NULL,
  country         CHAR(2) NOT NULL,
  created         TIMESTAMP NOT NULL DEFAULT NOW(),
  modified        TIMESTAMP NOT NULL DEFAULT NOW(),
  FOREIGN KEY (customer_id) REFERENCES customers(id)
);

CREATE INDEX IF NOT EXISTS created_idx  ON addresses (created DESC);
CREATE INDEX IF NOT EXISTS modified_idx ON addresses (modified DESC);

