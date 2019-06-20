CREATE FUNCTION is_leaf(p text)
RETURNS bool AS
$$
SELECT EXISTS (
  SELECT path FROM categories WHERE path = $1 AND lft = rgt - 1
)
$$
LANGUAGE 'sql' VOLATILE;

CREATE TABLE IF NOT EXISTS categories_products (
  id             SERIAL PRIMARY KEY,
  category_id    INTEGER NOT NULL,
  product_id     INTEGER NOT NULL,
  path           VARCHAR(1024) NOT NULL CHECK ( is_leaf(path) = TRUE ),
  sku            VARCHAR(64) NOT NULL,
  pri            INTEGER NOT NULL,
  created        TIMESTAMP NOT NULL DEFAULT NOW(),
  modified       TIMESTAMP NOT NULL DEFAULT NOW(),
  FOREIGN KEY (category_id, path) REFERENCES categories (id, path),
  FOREIGN KEY (product_id, sku) REFERENCES products (id, sku),
  UNIQUE (category_id, product_id)
);

CREATE INDEX IF NOT EXISTS idx_categories_products_pri ON categories_products (pri ASC);
CREATE INDEX IF NOT EXISTS idx_categories_products_created_desc ON categories_products (created DESC);
CREATE INDEX IF NOT EXISTS idx_categories_products_modified ON categories_products (modified DESC);
