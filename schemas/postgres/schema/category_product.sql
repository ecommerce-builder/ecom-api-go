CREATE FUNCTION is_leaf(p text)
RETURNS bool AS
$$
SELECT EXISTS (
  SELECT path FROM category WHERE path = $1 AND lft = rgt - 1
)
$$
LANGUAGE 'sql' VOLATILE;

CREATE TABLE IF NOT EXISTS category_product (
  id             SERIAL PRIMARY KEY,
  category_id    INTEGER NOT NULL,
  product_id     INTEGER NOT NULL,
  pri            INTEGER NOT NULL,
  created        TIMESTAMP NOT NULL DEFAULT NOW(),
  modified       TIMESTAMP NOT NULL DEFAULT NOW(),
  FOREIGN KEY (category_id) REFERENCES category (id),
  FOREIGN KEY (product_id) REFERENCES product (id),
  UNIQUE (category_id, product_id)
);

CREATE INDEX IF NOT EXISTS idx_category_product_pri ON category_product (pri ASC);
CREATE INDEX IF NOT EXISTS idx_category_product_created_desc ON category_product (created DESC);
CREATE INDEX IF NOT EXISTS idx_category_product_modified ON category_product (modified DESC);
