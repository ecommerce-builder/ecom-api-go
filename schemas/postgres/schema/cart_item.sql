CREATE TABLE IF NOT EXISTS cart_item (
  id          SERIAL PRIMARY KEY,
  uuid        UUID NOT NULL DEFAULT uuid_generate_v4(),
  cart_id     INTEGER,
  product_id  INTEGER,
  qty         SMALLINT NOT NULL CHECK (qty >= 1 AND qty < 10000),
  created     TIMESTAMP NOT NULL DEFAULT NOW(),
  modified    TIMESTAMP NOT NULL DEFAULT NOW(),
  UNIQUE (cart_id, product_id),
  FOREIGN KEY (cart_id) REFERENCES cart (id),
  FOREIGN KEY (product_id) REFERENCES product (id)
);
