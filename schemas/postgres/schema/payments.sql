DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'payment_t') THEN
        CREATE TYPE payment_t AS ENUM('stripe', 'paypal');
    END IF;
    
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'payment_status_t') THEN
        CREATE TYPE payment_status_t AS ENUM('success', 'failed', 'pending');
    END IF;
END$$;

CREATE TABLE IF NOT EXISTS payments (
  id            SERIAL PRIMARY KEY,
  uuid          UUID DEFAULT uuid_generate_v4() UNIQUE,
  order_id      INTEGER NOT NULL,
  typ           payment_t NOT NULL,
  status        payment_status_t NOT NULL,
  result        JSONB,
  created       TIMESTAMP NOT NULL DEFAULT NOW(),
  FOREIGN KEY (order_id) REFERENCES orders(id)
);
