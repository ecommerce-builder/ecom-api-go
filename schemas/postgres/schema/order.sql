CREATE TYPE order_status_t
  AS ENUM ('incomplete', 'completed');

CREATE TYPE order_payment_status_t
  AS ENUM ('unpaid', 'paid');

CREATE TABLE IF NOT EXISTS "order" (
  id              SERIAL PRIMARY KEY,
  uuid            UUID DEFAULT uuid_generate_v4() UNIQUE,
  usr_id          INTEGER NULL,
  status          order_status_t NOT NULL DEFAULT 'incomplete',
  payment         order_payment_status_t NOT NULL DEFAULT 'unpaid',
  contact_name    VARCHAR(512) NULL DEFAULT NULL,
  email           VARCHAR(512) NULL DEFAULT NULL,
  stripe_pi       VARCHAR(64) NULL DEFAULT NULL,
  billing_id      INTEGER NOT NULL,
  shipping_id     INTEGER NOT NULL,
  currency        CHAR(3) NOT NULL DEFAULT 'GBP',
  total_ex_vat    INTEGER NOT NULL CHECK (total_ex_vat >= 0),
  vat_total       INTEGER NOT NULL CHECK (vat_total >= 0),
  total_inc_vat   INTEGER NOT NULL CHECK (total_inc_vat >= 0 AND total_inc_vat = total_ex_vat + vat_total),
  created         TIMESTAMP NOT NULL DEFAULT NOW(),
  modified        TIMESTAMP NOT NULL DEFAULT NOW(),
  FOREIGN KEY (usr_id) REFERENCES usr (id),
  FOREIGN KEY (billing_id) REFERENCES order_address (id),
  FOREIGN KEY (shipping_id) REFERENCES order_address (id)
);

ALTER SEQUENCE order_id_seq RESTART WITH 100001;
