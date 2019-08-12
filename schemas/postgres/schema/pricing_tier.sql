CREATE TABLE IF NOT EXISTS pricing_tier (
  id            SERIAL PRIMARY KEY,
  uuid          UUID DEFAULT uuid_generate_v4() NOT NULL UNIQUE,
  tier_ref      VARCHAR(64) NOT NULL UNIQUE,
  title         VARCHAR(256) NOT NULL UNIQUE,
  description   VARCHAR(256) NULL,
  created       TIMESTAMP NOT NULL DEFAULT NOW(),
  modified      TIMESTAMP NOT NULL DEFAULT NOW()
);
