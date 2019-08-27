CREATE TABLE IF NOT EXISTS usr (
  id               SERIAL PRIMARY KEY,
  uuid             UUID DEFAULT uuid_generate_v4() UNIQUE,
  uid              VARCHAR(64) NOT NULL UNIQUE,
  price_list_id    INTEGER NOT NULL DEFAULT 1,
  role             VARCHAR(64) NOT NULL,
  email            VARCHAR(512) NOT NULL UNIQUE,
  firstname        VARCHAR(255) NOT NULL,
  lastname         VARCHAR(255) NOT NULL,
  created          TIMESTAMP NOT NULL DEFAULT NOW(),
  modified         TIMESTAMP NOT NULL DEFAULT NOW(),
  FOREIGN KEY (price_list_id) REFERENCES price_list (id)
);

CREATE INDEX IF NOT EXISTS idx_usr_role_asc ON usr (role ASC);
CREATE INDEX IF NOT EXISTS idx_usr_created_desc ON usr (created DESC);
CREATE INDEX IF NOT EXISTS idx_usr_modified ON usr (modified DESC);
