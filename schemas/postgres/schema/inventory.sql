CREATE TABLE IF NOT EXISTS inventory (
  id               SERIAL PRIMARY KEY,
  uuid             UUID DEFAULT uuid_generate_v4() UNIQUE,
  product_id       INTEGER NOT NULL,
  onhand           INTEGER CHECK (onhand > 0),
  created          TIMESTAMP NOT NULL DEFAULT NOW(),
  modified         TIMESTAMP NOT NULL DEFAULT NOW(),
  FOREIGN KEY (product_id) REFERENCES product (id)
);
