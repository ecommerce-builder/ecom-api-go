package postgres

import (
	"context"
	"database/sql"
	"time"

	"github.com/pkg/errors"
)

// ErrProductSetNotFound error
var ErrProductSetNotFound = errors.New("postgres: product set not found")

// ProductSetRow represents a single row in the product_set table.
type ProductSetRow struct {
	id       int
	UUID     string
	Created  time.Time
	Modified time.Time
}

// ProductSetItemJoinRow represents a single row in the product_set_list table.
type ProductSetItemJoinRow struct {
	id             int
	UUID           string
	productSetID   int
	ProductSetUUID string
	productID      int
	ProductUUID    string
	ProductPath    string
	ProductSKU     string
	Created        time.Time
	Modified       time.Time
}

// GetProductSetItems returns a list of product set items.
func (m *PgModel) GetProductSetItems(ctx context.Context, productSetUUID string) ([]*ProductSetItemJoinRow, error) {
	// 1. Check the product set exists
	q1 := "SELECT id FROM product_set WHERE uuid = $1"
	var productSetID int
	err := m.db.QueryRowContext(ctx, q1, productSetUUID).Scan(&productSetID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrProductSetNotFound
		}
		return nil, errors.Wrapf(err, "postgres: query row context failed for q1=%q", q1)
	}

	// 2. Get a list of all product set items in this product set.
	q2 := `
		SELECT
		  i.id, i.uuid, i.product_set_id, i.product_id,
		  p.uuid as product_uuid, p.path as product_path,
		  p.sku as product_sku,
		  i.created, i.modified
		FROM product_set_item AS i
		INNER JOIN product AS p
		  ON p.id = i.product_id
		WHERE product_set_id = $1
	`
	rows, err := m.db.QueryContext(ctx, q2, productSetID)
	if err != nil {
		return nil, errors.Wrapf(err, "postgres: m.db.QueryContext(ctx) failed")
	}
	defer rows.Close()

	productSetItems := make([]*ProductSetItemJoinRow, 0, 4)
	for rows.Next() {
		var p ProductSetItemJoinRow
		if err = rows.Scan(&p.id, &p.UUID, &p.productSetID, &p.productID, &p.ProductUUID, &p.ProductPath, &p.ProductSKU, &p.Created, &p.Modified); err != nil {
			return nil, errors.Wrap(err, "postgres: scan failed")
		}
		p.ProductSetUUID = productSetUUID
		productSetItems = append(productSetItems, &p)
	}
	if err = rows.Err(); err != nil {
		return nil, errors.Wrap(err, "postgres: rows.Err()")
	}

	return productSetItems, nil
}
