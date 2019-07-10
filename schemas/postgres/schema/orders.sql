CREATE TYPE order_status
  AS ENUM ('incomplete', 'completed');

CREATE TYPE order_payment_status
  AS ENUM ('unpaid', 'paid');

CREATE TYPE order_type
  AS ENUM ('guest', 'customer');

CREATE TABLE IF NOT EXISTS orders (
  id              SERIAL PRIMARY KEY,
  uuid            UUID DEFAULT uuid_generate_v4() UNIQUE,
  otype           order_type NOT NULL,
  status          order_status DEFAULT 'incomplete',
  payment         order_payment_status DEFAULT 'unpaid',
  customer_id     INTEGER NULL,
  customer_name   VARCHAR(512) NULL DEFAULT NULL,
  customer_email  VARCHAR(512) NULL DEFAULT NULL,
  ship_tb         BOOL DEFAULT false,
  billing         JSONB NOT NULL,
  shipping        JSONB,
  currency        CHAR(3) NOT NULL DEFAULT 'GBP',
  total_ex_vat    INTEGER NOT NULL CHECK (total_ex_vat >= 0),
  vat_total       INTEGER NOT NULL CHECK (vat_total >= 0),
  total_inc_vat   INTEGER NOT NULL CHECK (total_inc_vat >= 0 AND total_inc_vat = total_ex_vat + vat_total),
  created         TIMESTAMP NOT NULL DEFAULT NOW(),
  modified        TIMESTAMP NOT NULL DEFAULT NOW(),
  FOREIGN KEY (customer_id) REFERENCES customers(id)
);

ALTER SEQUENCE orders_id_seq RESTART WITH 100001;
