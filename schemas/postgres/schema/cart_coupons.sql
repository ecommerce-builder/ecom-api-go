CREATE TABLE IF NOT EXISTS cart_coupons (
  id             SERIAL PRIMARY KEY,
  uuid           UUID NOT NULL DEFAULT uuid_generate_v4(),
  coupon_id      INTEGER NOT NULL,
  cart_id        INTEGER NOT NULL,
  created        TIMESTAMP NOT NULL DEFAULT NOW(),
  modified       TIMESTAMP NOT NULL DEFAULT NOW(),
  UNIQUE (coupon_id, cart_id),
  FOREIGN KEY (coupon_id) REFERENCES coupons (id),
  FOREIGN KEY (cart_id) REFERENCES carts (id)
);
