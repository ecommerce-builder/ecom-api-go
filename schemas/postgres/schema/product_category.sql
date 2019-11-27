CREATE FUNCTION is_leaf_path(p text)
RETURNS bool AS
$$
SELECT EXISTS (
  SELECT path FROM category WHERE path = $1 AND lft = rgt - 1
)
$$
LANGUAGE 'sql' VOLATILE;

CREATE FUNCTION is_leaf(p integer)
RETURNS bool AS
$$
SELECT EXISTS (
  SELECT path FROM category WHERE id = $1 AND lft = rgt - 1
)
$$
LANGUAGE 'sql' VOLATILE;


CREATE TABLE IF NOT EXISTS product_category (
  id             SERIAL PRIMARY KEY,
  uuid           UUID DEFAULT uuid_generate_v4() NOT NULL UNIQUE,
  product_id     INTEGER NOT NULL,
  category_id    INTEGER NOT NULL CHECK ( is_leaf(category_id) = TRUE ),
  pri            INTEGER NOT NULL,
  created        TIMESTAMP NOT NULL DEFAULT NOW(),
  modified       TIMESTAMP NOT NULL DEFAULT NOW(),
  FOREIGN KEY (product_id) REFERENCES product (id),
  FOREIGN KEY (category_id) REFERENCES category (id),
  UNIQUE (product_id, category_id)
);

CREATE INDEX IF NOT EXISTS idx_product_category_pri ON product_category (pri ASC);
CREATE INDEX IF NOT EXISTS idx_product_category_created_desc ON product_category (created DESC);
CREATE INDEX IF NOT EXISTS idx_product_category_modified ON product_category (modified DESC);
