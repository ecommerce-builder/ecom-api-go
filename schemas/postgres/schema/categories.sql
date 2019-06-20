CREATE TABLE IF NOT EXISTS categories (
  id        SERIAL PRIMARY KEY,
  segment   VARCHAR(512) NOT NULL,
  path      VARCHAR(1024) NOT NULL,
  name      VARCHAR(1024) NOT NULL,
  lft       INTEGER NOT NULL UNIQUE,
  rgt       INTEGER NOT NULL UNIQUE,
  depth     INTEGER NOT NULL,
  created   TIMESTAMP NOT NULL DEFAULT NOW(),
  modified  TIMESTAMP NOT NULL DEFAULT NOW(),
  UNIQUE (lft, rgt),
  UNIQUE (id, path),
  UNIQUE (path)
);
