CREATE TABLE product_set_item (
  id             SERIAL PRIMARY KEY,
  uuid           UUID UNIQUE NOT NULL DEFAULT uuid_generate_v4(),
  product_set_id INTEGER NOT NULL,
  product_id     INTEGER NOT NULL,
  created        TIMESTAMP NOT NULL DEFAULT NOW(),
  modified       TIMESTAMP NOT NULL DEFAULT NOW(),
  UNIQUE (product_set_id, product_id),
  FOREIGN KEY    (product_set_id) REFERENCES product_set (id),
  FOREIGN KEY    (product_id) REFERENCES product (id)
);
