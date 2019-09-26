CREATE TABLE pp_assoc (
  id                      SERIAL PRIMARY KEY,
  uuid                    UUID UNIQUE NOT NULL DEFAULT uuid_generate_v4(),
  pp_assoc_group_id       INTEGER NOT NULL,
  product_from            INTEGER NOT NULL,
  product_to              INTEGER NOT NULL,
  created                 TIMESTAMP NOT NULL DEFAULT NOW(),
  modified                TIMESTAMP NOT NULL DEFAULT NOW(),
  FOREIGN KEY (pp_assoc_group_id) REFERENCES pp_assoc_group (id),
  FOREIGN KEY (product_from) REFERENCES product (id),
  FOREIGN KEY (product_from) REFERENCES product (id),
  UNIQUE (pp_assoc_group_id, product_from, product_to)
);
