CREATE TABLE IF NOT EXISTS cart_coupon (
  id             SERIAL PRIMARY KEY,
  uuid           UUID NOT NULL DEFAULT uuid_generate_v4(),
  cart_id        INTEGER NOT NULL,
  coupon_id      INTEGER NOT NULL,
  created        TIMESTAMP NOT NULL DEFAULT NOW(),
  modified       TIMESTAMP NOT NULL DEFAULT NOW(),
  UNIQUE (coupon_id, cart_id),
  FOREIGN KEY (cart_id) REFERENCES cart (id),
  FOREIGN KEY (coupon_id) REFERENCES coupon (id)
);
