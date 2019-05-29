package postgres

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/pkg/errors"
)

type CatalogProduct struct {
	ID        int
	CatalogID int
	ProductID int
	Path      string
	SKU       string
	Pri       int
	Created   time.Time
	Modified  time.Time
}

// CatalogProductAssoc maps products to leaf nodes in the catalogue hierarchy.
type CatalogProductAssoc struct {
	id        int
	catalogID int
	productID int
	Path      string
	SKU       string
	Pri       int
	Created   time.Time
	Modified  time.Time
}

// CreateCatalogProductAssoc links an existing product identified by sku
// to an existing leaf node of the catalog denoted by path.
func (m *PgModel) CreateCatalogProductAssoc(ctx context.Context, path, sku string) (*CatalogProduct, error) {
	query := `
		INSERT INTO catalog_products
			(catalog_id, product_id, path, sku, pri)
		VALUES (
			(SELECT id FROM catalog WHERE path = $1),
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
				FROM catalog_products
				WHERE path=$5
			)
		)
		RETURNING
			id, catalog_id, product_id, path, sku, pri, created, modified
	`
	cp := CatalogProduct{}
	err := m.db.QueryRowContext(ctx, query, path, sku, path, sku, path).Scan(&cp.ID, &cp.CatalogID, &cp.ProductID, &cp.Path, &cp.SKU, &cp.Pri, &cp.Created, &cp.Modified)
	if err != nil {
		return nil, errors.Wrapf(err, "model: query row context scan query=%q", query)
	}
	return &cp, nil
}

// BatchCreateCatalogProductAssocs inserts multiple catalog product
// associations using a transaction.
func (m *PgModel) BatchCreateCatalogProductAssocs(ctx context.Context, cpas map[string][]string) error {
	tx, err := m.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	query := "DELETE FROM catalog_products"
	_, err = tx.ExecContext(ctx, query)
	if err != nil {
		tx.Rollback()
		return errors.Wrapf(err, "model: delete catalog_products query=%q", query)
	}

	query = `
		INSERT INTO catalog_products
			(catalog_id, product_id, path, sku, pri)
		VALUES (
			(SELECT id FROM catalog WHERE path = $1),
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
				FROM catalog_products
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

// DeleteCatalogProductAssoc delete an existing catalog product association.
func (m *PgModel) DeleteCatalogProductAssoc(ctx context.Context, path, sku string) error {
	query := `
		DELETE FROM catalog_products
		WHERE path = $1 AND sku = $2
	`
	_, err := m.db.ExecContext(ctx, query, path, sku)
	if err != nil {
		return errors.Wrapf(err, "service: delete catalog product assoc path=%q sku=%q", path, sku)
	}
	return nil
}

// GetCatalogProductAssocs returns an Slice of catalogue to product
// associations.
func (m *PgModel) GetCatalogProductAssocs(ctx context.Context) ([]*CatalogProductAssoc, error) {
	query := `
		SELECT id, catalog_id, product_id, path, sku, pri, created, modified
		FROM catalog_products
		ORDER BY path, pri ASC
	`
	rows, err := m.db.QueryContext(ctx, query)
	if err != nil {
		return nil, errors.Wrapf(err, "model: query context query=%q", query)
	}
	defer rows.Close()
	cpas := make([]*CatalogProductAssoc, 0, 256)
	for rows.Next() {
		var n CatalogProductAssoc
		err = rows.Scan(&n.id, &n.catalogID, &n.productID, &n.Path, &n.SKU, &n.Pri, &n.Created, &n.Modified)
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

// HasCatalogProductAssocs returns true if any catalog product associations
// exist.
func (m *PgModel) HasCatalogProductAssocs(ctx context.Context) (bool, error) {
	query := "SELECT COUNT(*) AS count FROM catalog_products"
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

// UpdateCatalogProductAssocs update the catalog product associations.
func (m *PgModel) UpdateCatalogProductAssocs(ctx context.Context, cpo []*CatalogProductAssoc) error {
	tx, err := m.db.Begin()
	if err != nil {
		return err
	}

	stmt, err := tx.Prepare(`
		INSERT INTO catalog_products
			(catalog_id, product_id, path, sku, pri)
		VALUES (
			(SELECT id FROM catalog WHERE path = $1),
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

// DeleteCatalogProductAssocs delete all catalog product associations.
func (m *PgModel) DeleteCatalogProductAssocs(ctx context.Context) (affected int64, err error) {
	res, err := m.db.ExecContext(ctx, "DELETE FROM catalog_products")
	if err != nil {
		return -1, errors.Wrap(err, "assocs: delete catalog product assocs")
	}
	count, err := res.RowsAffected()
	if err != nil {
		return -1, errors.Wrap(err, "assocs: rows affected")
	}
	return count, nil
}
