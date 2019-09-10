CREATE TABLE product_set (
  id             SERIAL PRIMARY KEY,
  uuid           UUID UNIQUE NOT NULL DEFAULT uuid_generate_v4(),
  created        TIMESTAMP NOT NULL DEFAULT NOW(),
  modified       TIMESTAMP NOT NULL DEFAULT NOW()
);
