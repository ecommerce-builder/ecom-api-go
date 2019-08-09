CREATE TABLE IF NOT EXISTS product_images (
  id          SERIAL PRIMARY KEY,
  uuid        UUID DEFAULT uuid_generate_v4() NOT NULL UNIQUE,
  product_id  INTEGER NOT NULL,
  w           INTEGER NOT NULL CHECK (w > 0),
  h           INTEGER NOT NULL CHECK (h > 0),
  path        VARCHAR(4096) NOT NULL UNIQUE,
  typ         VARCHAR(64) NOT NULL DEFAULT 'image/jpeg',
  ori         BOOLEAN NOT NULL,
  up          BOOLEAN NOT NULL DEFAULT false,
  pri         INTEGER NOT NULL CHECK (pri > 0),
  size        INTEGER NOT NULL CHECK (size >= 0),
  q           INTEGER NOT NULL CHECK (q BETWEEN 1 AND 100),
  gsurl       VARCHAR(4096) NOT NULL UNIQUE,
  data        JSONB,
  created     TIMESTAMP NOT NULL DEFAULT NOW(),
  modified    TIMESTAMP NOT NULL DEFAULT NOW(),
  FOREIGN KEY (product_id) REFERENCES products (id)
);

CREATE INDEX IF NOT EXISTS pi_created_idx  ON product_images (created DESC);
CREATE INDEX IF NOT EXISTS pi_modified_idx ON product_images (modified DESC);
CREATE INDEX IF NOT EXISTS pi_w_idx  ON product_images (w ASC);
CREATE INDEX IF NOT EXISTS pi_h_idx ON product_images (h ASC);
CREATE INDEX IF NOT EXISTS pi_h_pri ON product_images (pri ASC);
CREATE INDEX IF NOT EXISTS pi_up ON product_images(up ASC);
CREATE INDEX IF NOT EXISTS pi_size_idx ON product_images (size DESC);
