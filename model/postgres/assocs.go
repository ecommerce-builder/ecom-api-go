package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"time"

	"github.com/pkg/errors"
)

// ErrProductCategoryExists error
var ErrProductCategoryExists = errors.New("postgres: product to category association exists")

// ErrProductCategoryNotFound error
var ErrProductCategoryNotFound = errors.New("postgres: product to category association not found")

// ProductCategoryRow represents a single row from the product_category table.
type ProductCategoryRow struct {
	id         int
	UUID       string
	productID  int
	categoryID int
	Pri        int
	Created    time.Time
	Modified   time.Time
}

// ProductCategoryBasicJoinRow contains a simple product_category row
// joined to the product and category table for uuids only.
type ProductCategoryBasicJoinRow struct {
	id           int
	UUID         string
	productID    int
	ProductUUID  string
	categoryID   int
	CategoryUUID string
	Pri          int
	Created      time.Time
	Modified     time.Time
}

// ProductCategoryJoinRow join row.
type ProductCategoryJoinRow struct {
	id              int
	UUID            string
	categoryID      int
	CategoryUUID    string
	CategoryPath    string
	productID       int
	ProductUUID     string
	ProductSKU      string
	ProductPath     string
	ProductName     string
	ProductCreated  time.Time
	ProductModified time.Time
	Pri             int
	Created         time.Time
	Modified        time.Time
}

// CreateProductCategoryRow represents a new row to add.
type CreateProductCategoryRow struct {
	CategoryUUID string
	ProductUUID  string
}

// AddProductCategory associates a product to a leaf category
func (m *PgModel) AddProductCategory(ctx context.Context, categoryUUID, productUUID string) (*ProductCategoryBasicJoinRow, error) {
	tx, err := m.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, errors.Wrap(err, "postgres: db.BeginTx")
	}

	// 1. Check if the category exists and that it's a leaf category
	q1 := "SELECT id, lft, rgt FROM category WHERE uuid = $1"
	var categoryID, lft, rgt int
	err = tx.QueryRowContext(ctx, q1, categoryUUID).Scan(&categoryID, &lft, &rgt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrCategoryNotFound
		}
		tx.Rollback()
		return nil, errors.Wrapf(err, "postgres: tx prepare for q1=%q", q1)
	}

	if lft != rgt-1 {
		return nil, ErrCategoryNotLeaf
	}

	// 2. Check if the product exists
	q2 := "SELECT id FROM product WHERE uuid = $1"
	var productID int
	err = tx.QueryRowContext(ctx, q2, productUUID).Scan(&productID)
	if err != nil {
		if err == sql.ErrNoRows {
			tx.Rollback()
			return nil, ErrProductNotFound
		}
		tx.Rollback()
		return nil, errors.Wrapf(err, "postgres: query row context failed for q2=%q", q2)
	}

	q3 := "SELECT EXISTS(SELECT 1 FROM product_category WHERE product_id = $1 AND category_id = $2) AS exists"
	var exists bool
	err = tx.QueryRowContext(ctx, q3, productID, categoryID).Scan(&exists)
	if err != nil {
		return nil, errors.Wrapf(err, "postgres: tx.QueryRowContext(ctx, q3=%q, productID=%d, categoryID=%d) failed", q3, productID, categoryID)
	}
	if exists {
		return nil, ErrProductCategoryExists
	}

	// 3. Link the product to the category
	q4 := `
		INSERT INTO product_category
		  (product_id, category_id, pri)
		VALUES ($1, $2,
			(
			SELECT
			  CASE WHEN COUNT(1) > 0
			  THEN MAX(pri)+10
			  ELSE 10
			END
			AS pri FROM product_category WHERE category_id = $3
			)
		)
		RETURNING
		  id, uuid,
		  product_id, (SELECT uuid AS product_uuid FROM product WHERE id = $4),
		  category_id, (SELECT uuid category_uuid FROM category WHERE id = $5),
		  pri, created, modified
		`
	row := tx.QueryRowContext(ctx, q4, productID, categoryID, categoryID, productID, categoryID)
	var pc ProductCategoryBasicJoinRow
	if err := row.Scan(&pc.id, &pc.UUID, &pc.productID, &pc.ProductUUID,
		&pc.categoryID, &pc.CategoryUUID, &pc.Pri, &pc.Created, &pc.Modified); err != nil {
		tx.Rollback()
		return nil, errors.Wrapf(err, "postgres: tx.QueryRowContext(ctx, q4, ...) failed q4=%q", q4)
	}

	if err = tx.Commit(); err != nil {
		return nil, errors.Wrap(err, "postgres: tx.Commit")
	}
	return &pc, nil
}

// DeleteProductCategory deletes a product_category row by uuid.
func (m *PgModel) DeleteProductCategory(ctx context.Context, productCategoryUUID string) error {
	tx, err := m.db.BeginTx(ctx, nil)
	if err != nil {
		return errors.Wrap(err, "postgres: db.BeginTx")
	}

	// 1. check if the product category exists
	q1 := "SELECT id FROM product_category WHERE uuid = $1"
	var productCategoryID int
	err = tx.QueryRowContext(ctx, q1, productCategoryUUID).Scan(&productCategoryID)
	if err != nil {
		if err == sql.ErrNoRows {
			tx.Rollback()
			return ErrProductCategoryNotFound
		}
		tx.Rollback()
		return errors.Wrapf(err, "postgres: query row context failed for q1=%q", q1)
	}

	// 2. delete it
	q2 := "DELETE FROM product_category WHERE id = $1"
	_, err = tx.ExecContext(ctx, q2, productCategoryID)
	if err != nil {
		return errors.Wrapf(err, "tx.ExecContext(ctx, q2, productCategoryID=%d)", productCategoryID)
	}

	if err = tx.Commit(); err != nil {
		return errors.Wrap(err, "postgres: tx.Commit")
	}
	return nil
}

// UpdateProductsCategories creates a batch of associations between a product and category.
func (m *PgModel) UpdateProductsCategories(ctx context.Context, cps []*CreateProductCategoryRow) ([]*ProductCategoryJoinRow, error) {
	tx, err := m.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, errors.Wrap(err, "postgres: db.BeginTx")
	}

	//
	// 1. Create a map of product id to product uuid
	//
	q1 := "SELECT id, uuid, path, sku, name FROM product"
	rows1, err := tx.QueryContext(ctx, q1)
	if err != nil {
		return nil, errors.Wrapf(err, "postgres: query context q1=%q", q1)
	}
	defer rows1.Close()

	type product struct {
		id   int
		uuid string
		path string
		sku  string
		name string
	}
	productMap := make(map[string]*product)
	for rows1.Next() {
		var p product
		err = rows1.Scan(&p.id, &p.uuid, &p.path, &p.sku, &p.name)
		if err != nil {
			return nil, errors.Wrapf(err, "postgres: scan failed")
		}
		productMap[p.uuid] = &p
	}
	if err = rows1.Err(); err != nil {
		return nil, errors.Wrapf(err, "postgres: rows.Err()")
	}

	//
	// 2. Create a map of category_id to category uuid
	//
	q2 := "SELECT id, uuid, path FROM category WHERE lft = rgt - 1"
	rows2, err := tx.QueryContext(ctx, q2)
	if err != nil {
		return nil, errors.Wrapf(err, "postgres: query context q2=%q", q2)
	}
	defer rows2.Close()

	type category struct {
		id   int
		uuid string
		path string
	}
	categoryMap := make(map[string]*category)
	for rows2.Next() {
		var c category
		err = rows2.Scan(&c.id, &c.uuid, &c.path)
		if err != nil {
			return nil, errors.Wrapf(err, "postgres: scan failed")
		}
		categoryMap[c.uuid] = &c
	}
	if err = rows2.Err(); err != nil {
		return nil, errors.Wrapf(err, "postgres: rows.Err()")
	}

	// Iterate the create products categories list passed in to this function.
	// Ensure that each product uuid and category uuid exists in the maps
	for _, c := range cps {
		if _, ok := productMap[c.ProductUUID]; !ok {
			return nil, ErrProductNotFound
		}
		if _, ok := categoryMap[c.CategoryUUID]; !ok {
			return nil, ErrLeafCategoryNotFound
		}
	}

	// 3. Delete the existing product category associations
	q3 := "DELETE FROM product_category"
	_, err = tx.ExecContext(ctx, q3)
	if err != nil {
		tx.Rollback()
		return nil, errors.Wrapf(err, "postgres: delete product_category q3=%q", q3)
	}

	q4 := `
		INSERT INTO product_category
		  (product_id, category_id, pri)
		VALUES ($1, $2,
			(
				SELECT
					CASE WHEN COUNT(1) > 0
					THEN MAX(pri)+10
					ELSE 10
				END
				AS pri
				FROM product_category
				WHERE category_id = $3
			)
		)
		RETURNING
		  id, uuid, product_id, category_id, pri, created, modified
	`
	stmt4, err := tx.PrepareContext(ctx, q4)
	if err != nil {
		tx.Rollback()
		return nil, errors.Wrapf(err, "postgres: tx prepare for q4=%q", q4)
	}
	defer stmt4.Close()

	productsCategories := make([]*ProductCategoryJoinRow, 0, len(cps))
	for _, r := range cps {
		product := productMap[r.ProductUUID]
		category := categoryMap[r.CategoryUUID]

		c := ProductCategoryJoinRow{}
		if err := stmt4.QueryRowContext(ctx, product.id, category.id, category.id).Scan(&c.id, &c.UUID, &c.productID, &c.categoryID, &c.Pri, &c.Created, &c.Modified); err != nil {
			if err == sql.ErrNoRows {
				tx.Rollback()
				return nil, ErrProductCategoryNotFound
			}
			tx.Rollback()
			return nil, errors.Wrapf(err, "postgres: stmt4.QueryRowContext(ctx, ...) failed q4=%q", q4)
		}
		c.ProductUUID = r.ProductUUID
		c.ProductPath = product.path
		c.ProductSKU = product.sku
		c.ProductName = product.name

		c.CategoryUUID = r.CategoryUUID
		c.CategoryPath = category.path
		productsCategories = append(productsCategories, &c)
	}

	if err = tx.Commit(); err != nil {
		return nil, errors.Wrap(err, "postgres: tx.Commit")
	}
	return productsCategories, nil
}

// CreateProductCategoryAssocs inserts multiple category product
// associations using a transaction.
func (m *PgModel) CreateProductCategoryAssocs(ctx context.Context, cpas map[string][]string) error {
	tx, err := m.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	q1 := "DELETE FROM product_category"
	_, err = tx.ExecContext(ctx, q1)
	if err != nil {
		tx.Rollback()
		return errors.Wrapf(err, "postgres: delete product_category query=%q", q1)
	}

	q2 := `
		INSERT INTO product_category
			(category_id, product_id, pri)
		VALUES (
			$1,
			$2,
			(
				SELECT
					CASE WHEN COUNT(1) > 0
					THEN MAX(pri) + 10
					ELSE 10
				END
				AS pri
				FROM product_category_
				WHERE category_id = $3
			)
		)
	`
	stmt2, err := tx.PrepareContext(ctx, q2)
	if err != nil {
		tx.Rollback()
		return errors.Wrapf(err, "tx prepare for query=%q", q2)
	}
	defer stmt2.Close()

	q3 := "SELECT id FROM category WHERE path = $1"
	stmt3, err := tx.PrepareContext(ctx, q3)
	if err != nil {
		tx.Rollback()
		return errors.Wrapf(err, "tx prepare for query=%q", q3)
	}

	q4 := "SELECT id FROM product WHERE uuid = $1"
	stmt4, err := tx.PrepareContext(ctx, q4)
	if err != nil {
		tx.Rollback()
		return errors.Wrapf(err, "tx prepare for query=%q", q4)
	}

	for path, pids := range cpas {
		var categoryID int
		err = stmt3.QueryRowContext(ctx, path).Scan(&categoryID)
		if err != nil {
			if err == sql.ErrNoRows {
				tx.Rollback()
				return ErrLeafCategoryNotFound
			}
			tx.Rollback()
			return errors.Wrapf(err, "query row context failed for query=%q", q3)
		}

		for _, pid := range pids {

			var productID int
			err = stmt4.QueryRowContext(ctx, pid).Scan(&productID)
			if err != nil {
				if err == sql.ErrNoRows {
					tx.Rollback()
					return ErrProductNotFound
				}
				tx.Rollback()
				return errors.Wrapf(err, "query row context failed for query=%q", q4)
			}

			if _, err := stmt2.ExecContext(ctx, categoryID, productID, categoryID); err != nil {
				tx.Rollback()
				fmt.Fprintf(os.Stderr, "%+v", err)
				return errors.Wrap(err, "stmt2 exec context")
			}
		}
	}

	if err = tx.Commit(); err != nil {
		return errors.Wrap(err, "tx.Commit")
	}
	return nil
}

// DeleteProductCategoryAssoc delete an existing catalog product association.
func (m *PgModel) DeleteProductCategoryAssoc(ctx context.Context, path, sku string) error {
	query := `
		DELETE FROM product_category
		WHERE path = $1 AND sku = $2
	`
	_, err := m.db.ExecContext(ctx, query, path, sku)
	if err != nil {
		return errors.Wrapf(err, "postgres: delete category product assoc path=%q sku=%q", path, sku)
	}
	return nil
}

// GetProductsCategories returns an list of product_category rows.
func (m *PgModel) GetProductsCategories(ctx context.Context) ([]*ProductCategoryJoinRow, error) {
	query := `
		SELECT
		  r.id, r.uuid, c.id as category_id, c.uuid as category_uuid, c.path as category_path,
		  product_id, p.uuid as product_uuid, p.sku, p.path as product_path,
		  p.name, pri, r.created, r.modified
		FROM
		  product_category AS r
		INNER JOIN category AS c
		  ON c.id = r.category_id
		INNER JOIN product AS p
		  ON p.id = r.product_id
		ORDER BY c.path, pri ASC;
	`
	rows, err := m.db.QueryContext(ctx, query)
	if err != nil {
		return nil, errors.Wrapf(err, "postgres: query context query=%q", query)
	}
	defer rows.Close()
	cpas := make([]*ProductCategoryJoinRow, 0, 256)
	for rows.Next() {
		var n ProductCategoryJoinRow
		err = rows.Scan(&n.id, &n.UUID, &n.categoryID, &n.CategoryUUID, &n.CategoryPath,
			&n.productID, &n.ProductUUID, &n.ProductSKU, &n.ProductPath,
			&n.ProductName, &n.Pri, &n.Created, &n.Modified)
		if err != nil {
			return nil, errors.Wrapf(err, "model: scan failed")
		}
		cpas = append(cpas, &n)
	}
	if err = rows.Err(); err != nil {
		return nil, errors.Wrapf(err, "postgres: rows.Err()")
	}
	return cpas, nil
}

// GetProductCategoryAssocsFull returns an Slice of catalog to product
// associations joined with products to including name.
func (m *PgModel) GetProductCategoryAssocsFull(ctx context.Context) ([]*ProductCategoryJoinRow, error) {
	query := `
		SELECT
		  c.id, category_id, product_id, p.uuid as product_uuid, t.path, p.path, p.sku, p.name,
		  p.created as product_created, p.modified as product_modified,
		  pri, c.created, c.modified
		FROM
		  product_category AS c
		INNER JOIN product AS p
		  ON p.id = c.product_id
		INNER JOIN category AS t
		  ON t.id = c.category_id
		ORDER BY p.path, pri ASC;
	`
	rows, err := m.db.QueryContext(ctx, query)
	if err != nil {
		return nil, errors.Wrapf(err, "postgres: query context query=%q", query)
	}
	defer rows.Close()
	cpas := make([]*ProductCategoryJoinRow, 0, 32)
	for rows.Next() {
		var n ProductCategoryJoinRow
		err = rows.Scan(&n.id, &n.categoryID, &n.productID, &n.ProductUUID, &n.CategoryPath, &n.ProductPath,
			&n.ProductSKU, &n.ProductName, &n.ProductCreated, &n.ProductModified, &n.Pri, &n.Created, &n.Modified)
		if err != nil {
			return nil, errors.Wrapf(err, "postgres: scan failed")
		}
		cpas = append(cpas, &n)
	}
	if err = rows.Err(); err != nil {
		return nil, errors.Wrapf(err, "postgres: rows.Err()")
	}
	return cpas, nil
}

// HasProductCategoryAssocs returns true if any catalog product associations
// exist.
func (m *PgModel) HasProductCategoryAssocs(ctx context.Context) (bool, error) {
	query := "SELECT COUNT(*) AS count FROM product_category"
	var count int
	err := m.db.QueryRowContext(ctx, query).Scan(&count)
	if err != nil {
		return false, errors.Wrapf(err, "postgres: query row context scan query=%q", query)
	}
	if count == 0 {
		return false, nil
	}
	return true, nil
}

// UpdateProductCategoryAssocs updates all entries in the product_category
// associations table.
// func (m *PgModel) UpdateProductCategoryAssocs(ctx context.Context, cpo []*ProductCategoryJoinRow) error {
// 	tx, err := m.db.BeginTx(ctx, nil)
// 	if err != nil {
// 		return err
// 	}
// 	stmt, err := tx.PrepareContext(ctx, `
// 		INSERT INTO product_category
// 			(category_id, product_id, path, sku, pri)
// 		VALUES (
// 			(SELECT id FROM category WHERE path = $1),
// 			(SELECT id FROM product WHERE sku = $2),
// 			$3,
// 			$4,
// 			$5
// 		)
// 	`)
// 	if err != nil {
// 		tx.Rollback()
// 		fmt.Fprintf(os.Stderr, "%v", err)
// 		return err
// 	}
// 	defer stmt.Close()
// 	for _, c := range cpo {
// 		if _, err := stmt.ExecContext(ctx, c.Path, c.SKU, c.Path, c.SKU, c.Pri); err != nil {
// 			tx.Rollback()
// 			fmt.Fprintf(os.Stderr, "%v", err)
// 			return err
// 		}
// 	}
// 	return tx.Commit()
// }

// PurgeProductsCategories
func (m *PgModel) PurgeProductsCategories(ctx context.Context) error {
	q1 := "DELETE FROM product_category"
	_, err := m.db.ExecContext(ctx, q1)
	if err != nil {
		return errors.Wrapf(err, "postgres: m.db.ExecContext(ctx, q1=%q)", q1)
	}
	return nil
}

// DeleteProductCategoryAssocs delete all category product
// associations effectly purging the catalog.
func (m *PgModel) DeleteProductCategoryAssocs(ctx context.Context) (affected int64, err error) {
	res, err := m.db.ExecContext(ctx, "DELETE FROM product_category")
	if err != nil {
		return -1, errors.Wrap(err, "postgres: delete product to category assocs")
	}
	count, err := res.RowsAffected()
	if err != nil {
		return -1, errors.Wrap(err, "postgres: rows affected")
	}
	return count, nil
}
