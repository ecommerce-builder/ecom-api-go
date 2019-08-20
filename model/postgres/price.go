package postgres

import (
	"context"
	"database/sql"
	"time"

	"github.com/pkg/errors"
)

// ErrPriceNotFound error.
var ErrPriceNotFound = errors.New("postgres: price not found")

// PriceRow represents a row in the price table
type PriceRow struct {
	id           int
	UUID         string
	productID    int
	priceListID  int
	Qty          int
	CurrencyCode string
	UnitPrice    int
	Created      time.Time
	Modified     time.Time
}

// PriceJoinRow represents a 3-way join between price,
// product and price_list.
type PriceJoinRow struct {
	id            int
	UUID          string
	productID     int
	ProductUUID   string
	SKU           string
	priceListID   *int
	PriceListUUID string
	PriceListCode string
	UnitPrice     int
	Created       time.Time
	Modified      time.Time
}

// GetPricesByPriceList returns a Price for a given product and price list.
func (m *PgModel) GetPricesByPriceList(ctx context.Context, productUUID, priceListUUID string) (*PriceJoinRow, error) {
	query := `
		SELECT
                  r.id, r.uuid, p.id, p.uuid AS product_uuid, p.sku, price_list_id, t.uuid AS price_list_uuid, t.code, unit_price, r.created, r.modified
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
	if err := m.db.QueryRowContext(ctx, query, productUUID, priceListUUID).Scan(&p.id, &p.UUID, &p.productID, &p.ProductUUID, &p.SKU, &p.priceListID, &p.PriceListUUID, &p.PriceListCode, &p.UnitPrice, &p.Created, &p.Modified); err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrPriceNotFound
		}
		return nil, errors.Wrapf(err, " m.db.QueryRowContext(ctx, %q, productUUID=%q, tierUUID=%q).Scan(...)", query, productUUID, priceListUUID)
	}
	return &p, nil
}

// GetPrices returns a list of prices for a given product.
func (m *PgModel) GetPrices(ctx context.Context, productUUID string) ([]*PriceJoinRow, error) {
	// TODO:
	// 1. Transaction
	// 2. check for missing product

	q2 := `
		SELECT
		  r.id, r.uuid AS uuid, p.id AS product_id, p.uuid as product_uuid, p.sku,
		  t.id as price_list_id, t.uuid as price_list_uuid, t.code, r.unit_price, r.created, r.modified
		FROM product AS p
		LEFT OUTER JOIN price AS r
		  ON p.id = r.product_id
		LEFT OUTER JOIN price_list AS t
		  ON t.id = r.price_list_id
		WHERE p.uuid = $1
		ORDER BY t.code
	`
	rows, err := m.db.QueryContext(ctx, q2, productUUID)
	if err != nil {
		return nil, errors.Wrapf(err, "m.db.QueryContext(ctx, q2=%q, productUUID=%q)", q2, productUUID)
	}
	defer rows.Close()
	pricing := make([]*PriceJoinRow, 0, 8)
	for rows.Next() {
		var p PriceJoinRow
		if err = rows.Scan(&p.id, &p.UUID, &p.productID, &p.ProductUUID, &p.SKU,
			&p.priceListID, &p.PriceListUUID, &p.PriceListCode, &p.UnitPrice, &p.Created, &p.Modified); err != nil {
			if err == sql.ErrNoRows {
				return nil, ErrPriceNotFound
			}
			return nil, errors.Wrap(err, "scan failed")
		}
		pricing = append(pricing, &p)
	}
	if err = rows.Err(); err != nil {
		return nil, errors.Wrap(err, "rows.Err()")
	}

	// If only one row is returned and the pricing tier id is nil
	// then there are no entries to return.
	if len(pricing) == 1 {
		if pricing[0].priceListID == nil {
			return nil, nil
		}
	}
	return pricing, nil
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
		if err = rows.Scan(&p.id, &p.priceListID, &p.productID, &p.UnitPrice, &p.Created, &p.Modified); err != nil {
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
		if err = rows.Scan(&p.id, &p.priceListID, &p.productID, &p.UnitPrice, &p.Created, &p.Modified); err != nil {
			return nil, errors.Wrap(err, "scan failed")
		}
		prices = append(prices, &p)
	}
	if err = rows.Err(); err != nil {
		return nil, errors.Wrap(err, "rows.Err()")
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
	if err != nil {
		if err == sql.ErrNoRows {
			tx.Rollback()
			return nil, ErrProductNotFound
		}
		tx.Rollback()
		return nil, errors.Wrapf(err, "query row context failed for query=%q", query)
	}

	query = "SELECT id FROM price_list WHERE uuid = $1"
	var priceListID int
	err = tx.QueryRowContext(ctx, query, priceListUUID).Scan(&priceListID)
	if err != nil {
		if err == sql.ErrNoRows {
			tx.Rollback()
			return nil, ErrPriceListNotFound
		}
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

	if err = tx.Commit(); err != nil {
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
	if err != nil {
		if err == sql.ErrNoRows {
			tx.Rollback()
			return ErrProductNotFound
		}
		tx.Rollback()
		return errors.Wrapf(err, "query row context failed for query=%q", query)
	}

	query = "SELECT id FROM price_list WHERE uuid = $1"
	var priceListID int
	err = tx.QueryRowContext(ctx, query, priceListUUID).Scan(&priceListID)
	if err != nil {
		if err == sql.ErrNoRows {
			tx.Rollback()
			return ErrPriceListNotFound
		}
		tx.Rollback()
		return errors.Wrapf(err, "query row context failed for query=%q", query)
	}

	query = "DELETE FROM price WHERE product_id = $1 AND price_list_id = $2"
	if _, err := m.db.ExecContext(ctx, query, productID, priceListID); err != nil {
		tx.Rollback()
		return errors.Wrapf(err, "exec context query=%q", query)
	}

	if err = tx.Commit(); err != nil {
		return errors.Wrap(err, "tx.Commit")
	}
	return nil
}
