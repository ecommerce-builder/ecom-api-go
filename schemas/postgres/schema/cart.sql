CREATE TABLE IF NOT EXISTS cart (
  id          SERIAL PRIMARY KEY,
  uuid        UUID UNIQUE NOT NULL DEFAULT uuid_generate_v4(),
  locked      BOOLEAN NOT NULL DEFAULT 'f',
  created     TIMESTAMP NOT NULL DEFAULT NOW(),
  modified    TIMESTAMP NOT NULL DEFAULT NOW()
);
