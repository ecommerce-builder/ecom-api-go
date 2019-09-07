CREATE TABLE IF NOT EXISTS usr_devkey (
  id           SERIAL PRIMARY KEY,
  uuid         UUID DEFAULT uuid_generate_v4() UNIQUE,
  usr_id       INTEGER NOT NULL,
  key          VARCHAR(256),
  hash         VARCHAR(256),
  created      TIMESTAMP NOT NULL DEFAULT NOW(),
  modified     TIMESTAMP NOT NULL DEFAULT NOW(),
  FOREIGN KEY (usr_id) REFERENCES usr (id),
  UNIQUE (usr_id, key)
);

CREATE INDEX IF NOT EXISTS idx_usr_devkey_apikey_asc ON usr_devkey (created DESC);
CREATE INDEX IF NOT EXISTS idx_usr_devkey_created_desc ON usr_devkey (created DESC);
CREATE INDEX IF NOT EXISTS idx_usr_devkey_modified ON usr_devkey (modified DESC);
