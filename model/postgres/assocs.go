package postgres

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/pkg/errors"
)

// CategoryProduct maps to a categories_product table row.
type CategoryProduct struct {
	id         int
	CategoryID int
	ProductID  int
	Path       string
	SKU        string
	Pri        int
	Created    time.Time
	Modified   time.Time
}

// CategoryProductAssoc maps products to leaf nodes in the catalog hierarchy.
type CategoryProductAssoc struct {
	id         int
	categoryID int
	productID  int
	Path       string
	SKU        string
	Pri        int
	Created    time.Time
	Modified   time.Time
}

// CategoryProductAssocFull maps products to leaf nodes in the catalog hierarchy.
type CategoryProductAssocFull struct {
	id           int
	categoryID   int
	productID    int
	CategoryPath string
	ProductPath  string
	SKU          string
	Name         string
	Pri          int
	Created      time.Time
	Modified     time.Time
}

// CreateCategoryProductAssoc links an existing product identified by sku
// to an existing leaf node of the catalog denoted by path.
func (m *PgModel) CreateCategoryProductAssoc(ctx context.Context, path, sku string) (*CategoryProduct, error) {
	query := `
		INSERT INTO categories_products
			(category_id, product_id, path, sku, pri)
		VALUES (
			(SELECT id FROM categories WHERE path = $1),
			(SELECT id FROM products WHERE sku = $2),
			$3,
			$4,
			(
				SELECT
					CASE WHEN COUNT(1) > 0
					THEN MAX(pri)+10
					ELSE 10
				END
				AS pri
				FROM categories_products
				WHERE path=$5
			)
		)
		RETURNING
			id, category_id, product_id, path, sku, pri, created, modified
	`
	cp := CategoryProduct{}
	row := m.db.QueryRowContext(ctx, query, path, sku, path, sku, path)
	if err := row.Scan(&cp.id, &cp.CategoryID, &cp.ProductID, &cp.Path, &cp.SKU, &cp.Pri, &cp.Created, &cp.Modified); err != nil {
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
	query := "DELETE FROM categories_products"
	_, err = tx.ExecContext(ctx, query)
	if err != nil {
		tx.Rollback()
		return errors.Wrapf(err, "model: delete categories_products query=%q", query)
	}
	query = `
		INSERT INTO categories_products
			(category_id, product_id, path, sku, pri)
		VALUES (
			(SELECT id FROM categories WHERE path = $1),
			(SELECT id FROM products WHERE sku = $2),
			$3,
			$4,
			(
				SELECT
					CASE WHEN COUNT(1) > 0
					THEN MAX(pri)+10
					ELSE 10
				END
				AS pri
				FROM categories_products
				WHERE path=$5
			)
		)
	`
	stmt, err := tx.PrepareContext(ctx, query)
	if err != nil {
		tx.Rollback()
		return errors.Wrapf(err, "tx prepare for query=%q", query)
	}
	defer stmt.Close()
	for path, skus := range cpas {
		for _, sku := range skus {
			if _, err := stmt.ExecContext(ctx, path, sku, path, sku, path); err != nil {
				tx.Rollback()
				fmt.Fprintf(os.Stderr, "%+v", err)
				return errors.Wrap(err, "stmt exec context")
			}
		}
	}
	return tx.Commit()
}

// DeleteCategoryProductAssoc delete an existing catalog product association.
func (m *PgModel) DeleteCategoryProductAssoc(ctx context.Context, path, sku string) error {
	query := `
		DELETE FROM categories_products
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
func (m *PgModel) GetCategoryProductAssocs(ctx context.Context) ([]*CategoryProductAssoc, error) {
	query := `
		SELECT id, category_id, product_id, path, sku, pri, created, modified
		FROM categories_products
		ORDER BY path, pri ASC
	`
	rows, err := m.db.QueryContext(ctx, query)
	if err != nil {
		return nil, errors.Wrapf(err, "model: query context query=%q", query)
	}
	defer rows.Close()
	cpas := make([]*CategoryProductAssoc, 0, 256)
	for rows.Next() {
		var n CategoryProductAssoc
		err = rows.Scan(&n.id, &n.categoryID, &n.productID, &n.Path, &n.SKU, &n.Pri, &n.Created, &n.Modified)
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
			C.id, category_id, product_id, C.path, P.path, C.sku, P.name,
			pri, C.created, C.modified
		FROM
			categories_products AS C,
			products AS P
		WHERE C.sku = P.sku
		ORDER BY C.path, pri ASC;
	`
	rows, err := m.db.QueryContext(ctx, query)
	if err != nil {
		return nil, errors.Wrapf(err, "model: query context query=%q", query)
	}
	defer rows.Close()
	cpas := make([]*CategoryProductAssocFull, 0, 32)
	for rows.Next() {
		var n CategoryProductAssocFull
		err = rows.Scan(&n.id, &n.categoryID, &n.productID, &n.CategoryPath, &n.ProductPath, &n.SKU, &n.Name, &n.Pri, &n.Created, &n.Modified)
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
	query := "SELECT COUNT(*) AS count FROM categories_products"
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

// UpdateCategoryProductAssocs updates all entries in the categories
// product associations table.
func (m *PgModel) UpdateCategoryProductAssocs(ctx context.Context, cpo []*CategoryProductAssoc) error {
	tx, err := m.db.Begin()
	if err != nil {
		return err
	}
	stmt, err := tx.PrepareContext(ctx, `
		INSERT INTO categories_products
			(category_id, product_id, path, sku, pri)
		VALUES (
			(SELECT id FROM categories WHERE path = $1),
			(SELECT id FROM products WHERE sku = $2),
			$3,
			$4,
			$5
		)
	`)
	if err != nil {
		tx.Rollback()
		fmt.Fprintf(os.Stderr, "%v", err)
		return err
	}
	defer stmt.Close()
	for _, c := range cpo {
		if _, err := stmt.ExecContext(ctx, c.Path, c.SKU, c.Path, c.SKU, c.Pri); err != nil {
			tx.Rollback()
			fmt.Fprintf(os.Stderr, "%v", err)
			return err
		}
	}
	return tx.Commit()
}

// DeleteCategoryProductAssocs delete all categories product
// associations effectly purging the catalog.
func (m *PgModel) DeleteCategoryProductAssocs(ctx context.Context) (affected int64, err error) {
	res, err := m.db.ExecContext(ctx, "DELETE FROM categories_products")
	if err != nil {
		return -1, errors.Wrap(err, "assocs: delete categories product assocs")
	}
	count, err := res.RowsAffected()
	if err != nil {
		return -1, errors.Wrap(err, "assocs: rows affected")
	}
	return count, nil
}
