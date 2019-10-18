CREATE TABLE IF NOT EXISTS price (
  id               SERIAL PRIMARY KEY,
  uuid             UUID DEFAULT uuid_generate_v4() NOT NULL UNIQUE,
  product_id       INTEGER,
  price_list_id    INTEGER,
  offer_id         INTEGER DEFAULT NULL,
  break            INTEGER NOT NULL CHECK (break >= 1),
  unit_price       INTEGER NOT NULL CHECK (unit_price >= 0),
  offer_price      INTEGER NULL CHECK (offer_price IS NULL OR offer_price >= 0),
  created          TIMESTAMP NOT NULL DEFAULT NOW(),
  modified         TIMESTAMP NOT NULL DEFAULT NOW(),
  FOREIGN KEY (product_id) REFERENCES product (id),
  FOREIGN KEY (price_list_id) REFERENCES price_list (id),
  UNIQUE (price_list_id, product_id, break)
);
