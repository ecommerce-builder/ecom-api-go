CREATE TABLE IF NOT EXISTS cart_offer (
  id          SERIAL PRIMARY KEY,
  uuid        UUID NOT NULL DEFAULT uuid_generate_v4(),
  cart_id     INTEGER,
  offer_id    INTEGER,
  created     TIMESTAMP NOT NULL DEFAULT NOW(),
  modified    TIMESTAMP NOT NULL DEFAULT NOW(),
  UNIQUE (cart_id, offer_id),
  FOREIGN KEY (cart_id) REFERENCES cart (id),
  FOREIGN KEY (offer_id) REFERENCES offer (id)
);
