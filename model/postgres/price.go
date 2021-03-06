package postgres

import (
	"context"
	"database/sql"
	"strings"
	"time"

	"github.com/pkg/errors"
)

// ErrPriceNotFound error.
var ErrPriceNotFound = errors.New("postgres: price not found")

// PriceRow represents a row in the price table
type PriceRow struct {
	id          int
	UUID        string
	productID   int
	priceListID int
	offerID     int
	Break       int
	UnitPrice   int
	Created     time.Time
	Modified    time.Time
}

// PriceJoinRow represents a 3-way join between price,
// product and price_list.
type PriceJoinRow struct {
	id            int
	UUID          string
	productID     int
	ProductUUID   string
	ProductPath   string
	ProductSKU    string
	priceListID   int
	PriceListUUID string
	PriceListCode string
	Break         int
	UnitPrice     int
	Created       time.Time
	Modified      time.Time
}

// CreatePrice is a container for break and unit price pairs.
type CreatePrice struct {
	Break     int
	UnitPrice int
}

// GetPricesByPriceList returns a Price for a given product and price list.
func (m *PgModel) GetPricesByPriceList(ctx context.Context, productUUID, priceListUUID string) (*PriceJoinRow, error) {
	query := `
		SELECT
		  r.id, r.uuid, p.id, p.uuid AS product_uuid, p.path, p.sku, price_list_id,
		  t.uuid AS price_list_uuid, t.code, unit_price, r.created, r.modified
                FROM
                  price r
		JOIN product p
		  ON p.id = r.product_id
		JOIN price_list t
		  ON t.id = r.price_list_id
                WHERE
                  p.uuid = $1 AND t.uuid = $2;
	`
	p := PriceJoinRow{}
	err := m.db.QueryRowContext(ctx, query, productUUID, priceListUUID).Scan(&p.id, &p.UUID, &p.productID, &p.ProductUUID, &p.ProductPath, &p.ProductSKU, &p.priceListID, &p.PriceListUUID, &p.PriceListCode, &p.UnitPrice, &p.Created, &p.Modified)
	if err == sql.ErrNoRows {
		return nil, ErrPriceNotFound
	}
	if err != nil {
		return nil, errors.Wrapf(err, " m.db.QueryRowContext(ctx, %q, productUUID=%q, tierUUID=%q).Scan(...)", query, productUUID, priceListUUID)
	}
	return &p, nil
}

// GetPrices returns a list of prices for a given product.
func (m *PgModel) GetPrices(ctx context.Context, productUUID, priceListUUID string) ([]*PriceJoinRow, error) {
	tx, err := m.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, errors.Wrap(err, "postgres: db.BeginTx")
	}

	// 1. Check product exists (if it's provided)
	var productID int
	if productUUID != "" {
		q1 := "SELECT id FROM product WHERE uuid = $1"
		err = tx.QueryRowContext(ctx, q1, productUUID).Scan(&productID)
		if err == sql.ErrNoRows {
			tx.Rollback()
			return nil, ErrProductNotFound
		}
		if err != nil {
			tx.Rollback()
			return nil, errors.Wrapf(err, "postgres: query row context failed for q1=%q", q1)
		}
	}

	// 2. Check price list exists (if it's provided)
	var priceListID int
	if priceListUUID != "" {
		q2 := "SELECT id FROM price_list WHERE uuid = $1"
		err = tx.QueryRowContext(ctx, q2, priceListUUID).Scan(&priceListID)
		if err == sql.ErrNoRows {
			tx.Rollback()
			return nil, ErrPriceListNotFound
		}
		if err != nil {
			tx.Rollback()
			return nil, errors.Wrapf(err, "postgres: query row context failed for q2=%q", q2)
		}
	}

	q3 := `
		SELECT
		  r.id, r.uuid AS uuid, p.id AS product_id, p.uuid as product_uuid, p.path, p.sku,
		  t.id as price_list_id, t.uuid as price_list_uuid, t.code,
		  r.unit_price, r.break, r.created, r.modified
		FROM product AS p
		INNER JOIN price AS r
		  ON p.id = r.product_id
		INNER JOIN price_list AS t
		  ON t.id = r.price_list_id
		%WHERECLAUSE%
		ORDER BY t.code
	`
	var rows *sql.Rows
	if productID > 0 && priceListID > 0 {
		q3 = strings.Replace(q3, "%WHERECLAUSE%", "WHERE p.id = $1 AND price_list_id = $2", 1)
		rows, err = tx.QueryContext(ctx, q3, productID, priceListID)
		if err != nil {
			return nil, errors.Wrapf(err, "postgres: m.db.QueryContext(ctx, q3=%q, productID=%q, priceListID=%q)", q3, productID, priceListID)
		}
	} else if productID > 0 {
		q3 = strings.Replace(q3, "%WHERECLAUSE%", "WHERE p.id = $1", 1)
		rows, err = tx.QueryContext(ctx, q3, productID)
		if err != nil {
			return nil, errors.Wrapf(err, "postgres: m.db.QueryContext(ctx, q3=%q, productID=%q)", q3, productID)
		}
	} else if priceListID > 0 {
		q3 = strings.Replace(q3, "%WHERECLAUSE%", "WHERE price_list_id = $1", 1)
		rows, err = tx.QueryContext(ctx, q3, priceListID)
		if err != nil {
			return nil, errors.Wrapf(err, "postgres: m.db.QueryContext(ctx, q3=%q, priceListID=%q)", q3, priceListID)
		}
	} else {
		q3 = strings.Replace(q3, "%WHERECLAUSE%", "", 1)
		rows, err = tx.QueryContext(ctx, q3)
		if err != nil {
			return nil, errors.Wrapf(err, "postgres: m.db.QueryContext(ctx, q3=%q)", q3)
		}
	}
	defer rows.Close()

	prices := make([]*PriceJoinRow, 0, 8)
	for rows.Next() {
		var p PriceJoinRow
		err = rows.Scan(&p.id, &p.UUID, &p.productID, &p.ProductUUID, &p.ProductPath, &p.ProductSKU,
			&p.priceListID, &p.PriceListUUID, &p.PriceListCode,
			&p.UnitPrice, &p.Break, &p.Created, &p.Modified)
		if err == sql.ErrNoRows {
			return nil, ErrPriceNotFound
		}
		if err != nil {
			return nil, errors.Wrap(err, "postgres: scan failed")
		}
		prices = append(prices, &p)
	}
	if err := rows.Err(); err != nil {
		return nil, errors.Wrap(err, "postgres: rows.Err()")
	}

	if err := tx.Commit(); err != nil {
		return nil, errors.Wrap(err, "postgres: tx.Commit")
	}
	return prices, nil
}

// TODO: fix this up
//

// GetProductPrices returns a list of prices for a given product.
func (m *PgModel) GetProductPrices(ctx context.Context, productUUID string) ([]*PriceRow, error) {
	query := `
		SELECT id, price_list_id, product_id, unit_price, created, modified
		FROM price
		WHERE sku = $1
		ORDER BY id ASC
	`
	rows, err := m.db.QueryContext(ctx, query, productUUID)
	if err != nil {
		return nil, errors.Wrapf(err, "m.db.QueryContext(ctx, %q, productUUID%q)", query, productUUID)
	}
	defer rows.Close()
	pricing := make([]*PriceRow, 0, 8)
	for rows.Next() {
		var p PriceRow
		if err := rows.Scan(&p.id, &p.priceListID, &p.productID, &p.UnitPrice, &p.Created, &p.Modified); err != nil {
			return nil, errors.Wrap(err, "scan failed")
		}
		pricing = append(pricing, &p)
	}
	if err = rows.Err(); err != nil {
		return nil, errors.Wrap(err, "rows.Err()")
	}
	return pricing, nil
}

// GetProductPriceByPriceList returns a list of Price items for a given price list id.
func (m *PgModel) GetProductPriceByPriceList(ctx context.Context, priceListUUID string) ([]*PriceRow, error) {
	query := `
		SELECT
		  id, uuid, price_list_id, product_id, unit_price, created, modified
		FROM price
		WHERE uuid = $1
	`
	rows, err := m.db.QueryContext(ctx, query, priceListUUID)
	if err != nil {
		return nil, errors.Wrapf(err, "m.db.QueryContext(ctx, query=%q, priceListID=%q)", query, priceListUUID)
	}
	defer rows.Close()

	prices := make([]*PriceRow, 0, 8)
	for rows.Next() {
		var p PriceRow
		if err := rows.Scan(&p.id, &p.priceListID, &p.productID, &p.UnitPrice, &p.Created, &p.Modified); err != nil {
			return nil, errors.Wrap(err, "scan failed")
		}
		prices = append(prices, &p)
	}
	if err := rows.Err(); err != nil {
		return nil, errors.Wrap(err, "rows.Err()")
	}
	return prices, nil
}

// UpdatePrices updates a batch of prices.
func (m *PgModel) UpdatePrices(ctx context.Context, productUUID, priceListUUID string, createPrices []*CreatePrice) ([]*PriceJoinRow, error) {
	tx, err := m.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, errors.Wrap(err, "postgres: db.BeginTx")
	}

	// 1. Check the product exists
	q1 := "SELECT id, path, sku FROM product WHERE uuid = $1"
	var productID int
	var productPath string
	var productSKU string
	err = tx.QueryRowContext(ctx, q1, productUUID).Scan(&productID, &productPath, &productSKU)
	if err == sql.ErrNoRows {
		tx.Rollback()
		return nil, ErrProductNotFound
	}
	if err != nil {
		tx.Rollback()
		return nil, errors.Wrapf(err, "postgres: query row context failed for q1=%q", q1)
	}

	// 2. Check the product list exists
	q2 := "SELECT id, code FROM price_list WHERE uuid = $1"
	var priceListID int
	var priceListCode string
	err = tx.QueryRowContext(ctx, q2, priceListUUID).Scan(&priceListID, &priceListCode)
	if err == sql.ErrNoRows {
		tx.Rollback()
		return nil, ErrPriceListNotFound
	}
	if err != nil {
		tx.Rollback()
		return nil, errors.Wrapf(err, "postgres: query row context failed for q2=%q", q2)
	}

	// 3. Delete the old prices for this product and price_list
	q3 := "DELETE FROM price WHERE product_id = $1 AND price_list_id = $2"
	if _, err := tx.ExecContext(ctx, q3, productID, priceListID); err != nil {
		tx.Rollback()
		return nil, errors.Wrapf(err, "exec context q3=%q", q3)
	}

	// 4. Replace the prices
	q4 := `
		INSERT INTO price
		  (product_id, price_list_id, break, unit_price)
		VALUES
		  ($1, $2, $3, $4)
		RETURNING
		  id, uuid, product_id, price_list_id, break, unit_price, created, modified
	`
	stmt4, err := tx.PrepareContext(ctx, q4)
	if err != nil {
		tx.Rollback()
		return nil, errors.Wrapf(err, "postgres: tx prepare for query=%q", q4)
	}
	defer stmt4.Close()

	prices := make([]*PriceJoinRow, 0, 2)
	for _, cp := range createPrices {
		var p PriceJoinRow
		row := stmt4.QueryRowContext(ctx, productID, priceListID, cp.Break, cp.UnitPrice)
		if err := row.Scan(&p.id, &p.UUID, &p.productID, &priceListID, &p.Break, &p.UnitPrice, &p.Created, &p.Modified); err != nil {
			tx.Rollback()
			return nil, errors.Wrapf(err, "postgres: row.Scan failed")
		}
		p.ProductUUID = productUUID
		p.ProductPath = productPath
		p.ProductSKU = productSKU
		p.PriceListUUID = priceListUUID
		p.PriceListCode = priceListCode
		prices = append(prices, &p)
	}

	if err := tx.Commit(); err != nil {
		return nil, errors.Wrap(err, "postgres: tx.Commit")
	}
	return prices, nil
}

// UpdatePrice updates the product price with the new unit price
// by product uuid and price list uuid.
func (m *PgModel) UpdatePrice(ctx context.Context, productUUID, priceListUUID string, unitPrice float64) (*PriceJoinRow, error) {
	tx, err := m.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, errors.Wrap(err, "db.BeginTx")
	}

	query := "SELECT id FROM product WHERE uuid = $1"
	var productID int
	err = tx.QueryRowContext(ctx, query, productUUID).Scan(&productID)
	if err == sql.ErrNoRows {
		tx.Rollback()
		return nil, ErrProductNotFound
	}
	if err != nil {
		tx.Rollback()
		return nil, errors.Wrapf(err, "query row context failed for query=%q", query)
	}

	query = "SELECT id FROM price_list WHERE uuid = $1"
	var priceListID int
	err = tx.QueryRowContext(ctx, query, priceListUUID).Scan(&priceListID)
	if err == sql.ErrNoRows {
		tx.Rollback()
		return nil, ErrPriceListNotFound
	}
	if err != nil {
		tx.Rollback()
		return nil, errors.Wrapf(err, "query row context failed for query=%q", query)
	}

	query = `
		UPDATE price
		SET
		  unit_price = $1, modified = NOW()
		WHERE
		  product_id = $2 AND price_list_id = $3
		RETURNING
		  id, uuid, price_list_id, code as (SELECT code FROM price_list WHERE id = $4) product_id, unit_price, created, modified
	`
	p := PriceJoinRow{}
	if err := tx.QueryRowContext(ctx, query, unitPrice, productID, priceListID, priceListID).Scan(&p.id, &p.UUID, &p.priceListID, &p.productID, &p.UnitPrice, &p.Created, &p.Modified); err != nil {
		return nil, errors.Wrap(err, "UpdateTierPricing QueryRowContext failed")
	}

	if err := tx.Commit(); err != nil {
		return nil, errors.Wrap(err, "tx.Commit")
	}
	return &p, nil
}

// DeleteProductPrices deletes a tier pricing by product uuid and pricing tier uuid.
func (m *PgModel) DeleteProductPrices(ctx context.Context, productUUID, priceListUUID string) error {
	tx, err := m.db.BeginTx(ctx, nil)
	if err != nil {
		return errors.Wrap(err, "db.BeginTx")
	}

	query := "SELECT id FROM product WHERE uuid = $1"
	var productID int
	err = tx.QueryRowContext(ctx, query, productUUID).Scan(&productID)
	if err == sql.ErrNoRows {
		tx.Rollback()
		return ErrProductNotFound
	}
	if err != nil {
		tx.Rollback()
		return errors.Wrapf(err, "query row context failed for query=%q", query)
	}

	query = "SELECT id FROM price_list WHERE uuid = $1"
	var priceListID int
	err = tx.QueryRowContext(ctx, query, priceListUUID).Scan(&priceListID)
	if err == sql.ErrNoRows {
		tx.Rollback()
		return ErrPriceListNotFound
	}
	if err != nil {
		tx.Rollback()
		return errors.Wrapf(err, "query row context failed for query=%q", query)
	}

	query = "DELETE FROM price WHERE product_id = $1 AND price_list_id = $2"
	if _, err := m.db.ExecContext(ctx, query, productID, priceListID); err != nil {
		tx.Rollback()
		return errors.Wrapf(err, "exec context query=%q", query)
	}

	if err := tx.Commit(); err != nil {
		return errors.Wrap(err, "tx.Commit")
	}
	return nil
}
