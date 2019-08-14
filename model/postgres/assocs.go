package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"time"

	"github.com/pkg/errors"
)

// CategoryProductRow represents a single row from the category_product table.
type CategoryProductRow struct {
	id         int
	categoryID int
	productID  int
	Pri        int
	Created    time.Time
	Modified   time.Time
}

// CategoryProductJoinRow maps products to leaf nodes in the catalog hierarchy.
type CategoryProductJoinRow struct {
	id              int
	categoryID      int
	CategoryUUID    string
	CategoryPath    string
	productID       int
	ProductUUID     string
	ProductSKU      string
	ProductEAN      string
	ProductPath     string
	ProductName     string
	ProductCreated  time.Time
	ProductModified time.Time
	Pri             int
	Created         time.Time
	Modified        time.Time
}

// CategoryProductAssocFull maps products to leaf nodes in the catalog hierarchy.
type CategoryProductAssocFull struct {
	id              int
	categoryID      int
	productID       int
	ProductUUID     string
	CategoryPath    string
	ProductPath     string
	SKU             string
	EAN             string
	Name            string
	ProductCreated  time.Time
	ProductModified time.Time
	Pri             int
	Created         time.Time
	Modified        time.Time
}

// CreateCategoryProductAssoc links an existing product identified by sku
// to an existing leaf node of the catalog denoted by path.
func (m *PgModel) CreateCategoryProductAssoc(ctx context.Context, path, sku string) (*CategoryProductRow, error) {
	query := `
		INSERT INTO category_product
			(category_id, product_id, pri)
		VALUES (
			(SELECT id FROM category WHERE path = $1),
			(SELECT id FROM product WHERE sku = $2),
			$3,
			$4,
			(
				SELECT
					CASE WHEN COUNT(1) > 0
					THEN MAX(pri)+10
					ELSE 10
				END
				AS pri
				FROM category_product
				WHERE path = $5
			)
		)
		RETURNING
		  id, category_id, product_id, pri, created, modified
	`
	cp := CategoryProductRow{}
	row := m.db.QueryRowContext(ctx, query, path, sku, path, sku, path)
	if err := row.Scan(&cp.id, &cp.categoryID, &cp.productID, &cp.Pri, &cp.Created, &cp.Modified); err != nil {
		return nil, errors.Wrapf(err, "model: query row context scan query=%q", query)
	}
	return &cp, nil
}

// CreateCategoryProductAssocs inserts multiple category product
// associations using a transaction.
func (m *PgModel) CreateCategoryProductAssocs(ctx context.Context, cpas map[string][]string) error {
	tx, err := m.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	q1 := "DELETE FROM category_product"
	_, err = tx.ExecContext(ctx, q1)
	if err != nil {
		tx.Rollback()
		return errors.Wrapf(err, "model: delete category_product query=%q", q1)
	}

	q2 := `
		INSERT INTO category_product
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
				FROM category_product
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
				return ErrCategoryNotFound
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

// DeleteCategoryProductAssoc delete an existing catalog product association.
func (m *PgModel) DeleteCategoryProductAssoc(ctx context.Context, path, sku string) error {
	query := `
		DELETE FROM category_product
		WHERE path = $1 AND sku = $2
	`
	_, err := m.db.ExecContext(ctx, query, path, sku)
	if err != nil {
		return errors.Wrapf(err, "service: delete category product assoc path=%q sku=%q", path, sku)
	}
	return nil
}

// GetCategoryProductAssocs returns an Slice of catalog to product
// associations.
func (m *PgModel) GetCategoryProductAssocs(ctx context.Context) ([]*CategoryProductJoinRow, error) {
	query := `
		SELECT
		  c.id, category_id, c.uuid as category_uuid, c.path as category_path, product_id, p.uuid as product_uuid, p.sku, p.ean, p.path as product_path, p.name, pri, c.created, c.modified
		FROM
		  category_product AS r
		INNER JOIN category AS c
		  ON c.id = r.category_id
		INNER JOIN product AS p
		  ON p.id = r.product_id
		ORDER BY c.path, pri ASC;
	`
	rows, err := m.db.QueryContext(ctx, query)
	if err != nil {
		return nil, errors.Wrapf(err, "model: query context query=%q", query)
	}
	defer rows.Close()
	cpas := make([]*CategoryProductJoinRow, 0, 256)
	for rows.Next() {
		var n CategoryProductJoinRow
		err = rows.Scan(&n.id, &n.categoryID, &n.CategoryUUID, &n.CategoryPath, &n.productID, &n.ProductUUID, &n.ProductSKU, &n.ProductEAN, &n.ProductPath, &n.ProductSKU, &n.Pri, &n.Created, &n.Modified)
		if err != nil {
			return nil, errors.Wrapf(err, "model: scan failed")
		}
		cpas = append(cpas, &n)
	}
	if err = rows.Err(); err != nil {
		return nil, errors.Wrapf(err, "model: rows.Err()")
	}
	return cpas, nil
}

// GetCategoryProductAssocsFull returns an Slice of catalog to product
// associations joined with products to including name.
func (m *PgModel) GetCategoryProductAssocsFull(ctx context.Context) ([]*CategoryProductAssocFull, error) {
	query := `
		SELECT
		  c.id, category_id, product_id, p.uuid as product_uuid, t.path, p.path, p.sku, p.ean, p.name,
		  p.created as product_created, p.modified as product_modified,
		  pri, c.created, c.modified
		FROM
		  category_product AS c
		INNER JOIN product AS p
		  ON p.id = c.product_id
		INNER JOIN category AS t
		  ON t.id = c.category_id
		ORDER BY p.path, pri ASC;
	`
	rows, err := m.db.QueryContext(ctx, query)
	if err != nil {
		return nil, errors.Wrapf(err, "model: query context query=%q", query)
	}
	defer rows.Close()
	cpas := make([]*CategoryProductAssocFull, 0, 32)
	for rows.Next() {
		var n CategoryProductAssocFull
		err = rows.Scan(&n.id, &n.categoryID, &n.productID, &n.ProductUUID, &n.CategoryPath, &n.ProductPath,
			&n.SKU, &n.EAN, &n.Name, &n.ProductCreated, &n.ProductModified, &n.Pri, &n.Created, &n.Modified)
		if err != nil {
			return nil, errors.Wrapf(err, "model: scan failed")
		}
		cpas = append(cpas, &n)
	}
	if err = rows.Err(); err != nil {
		return nil, errors.Wrapf(err, "model: rows.Err()")
	}
	return cpas, nil
}

// HasCategoryProductAssocs returns true if any catalog product associations
// exist.
func (m *PgModel) HasCategoryProductAssocs(ctx context.Context) (bool, error) {
	query := "SELECT COUNT(*) AS count FROM category_product"
	var count int
	err := m.db.QueryRowContext(ctx, query).Scan(&count)
	if err != nil {
		return false, errors.Wrapf(err, "query row context scan query=%q", query)
	}
	if count == 0 {
		return false, nil
	}
	return true, nil
}

// UpdateCategoryProductAssocs updates all entries in the category_product
// associations table.
// func (m *PgModel) UpdateCategoryProductAssocs(ctx context.Context, cpo []*CategoryProductJoinRow) error {
// 	tx, err := m.db.BeginTx(ctx, nil)
// 	if err != nil {
// 		return err
// 	}
// 	stmt, err := tx.PrepareContext(ctx, `
// 		INSERT INTO category_product
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

// DeleteCategoryProductAssocs delete all category product
// associations effectly purging the catalog.
func (m *PgModel) DeleteCategoryProductAssocs(ctx context.Context) (affected int64, err error) {
	res, err := m.db.ExecContext(ctx, "DELETE FROM category_product")
	if err != nil {
		return -1, errors.Wrap(err, "assocs: delete category product assocs")
	}
	count, err := res.RowsAffected()
	if err != nil {
		return -1, errors.Wrap(err, "assocs: rows affected")
	}
	return count, nil
}
