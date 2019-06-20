CREATE TABLE IF NOT EXISTS orders (
  id           SERIAL PRIMARY KEY,
  uuid         UUID DEFAULT uuid_generate_v4() UNIQUE,
  customer_id  INTEGER,
  ship_tb      BOOL DEFAULT false,
  billing      JSONB NOT NULL,
  shipping     JSONB,
  total        NUMERIC(8, 4) NOT NULL CHECK (total >= 0.0000),
  created      TIMESTAMP NOT NULL DEFAULT NOW(),
  FOREIGN KEY (customer_id) REFERENCES customers(id)
);

ALTER SEQUENCE orders_id_seq RESTART WITH 100001;
