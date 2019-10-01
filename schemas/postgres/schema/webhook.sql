CREATE TABLE IF NOT EXISTS webhook (
  id           SERIAL PRIMARY KEY,
  uuid         UUID UNIQUE NOT NULL DEFAULT uuid_generate_v4(),
  signing_key  VARCHAR(256) NOT NULL,
  url          VARCHAR(1024) NOT NULL UNIQUE,
  events       VARCHAR(128)[] NOT NULL,
  enabled      BOOLEAN NOT NULL DEFAULT true,
  created      TIMESTAMP NOT NULL DEFAULT NOW(),
  modified     TIMESTAMP NOT NULL DEFAULT NOW()
);
