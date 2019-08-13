package postgres

import (
	"context"
	"database/sql"
	"time"

	"github.com/pkg/errors"
)

// ErrProductPricingNotFound error.
var ErrProductPricingNotFound = errors.New("product pricing not found")

// ErrPricingTierNotFound error
var ErrPricingTierNotFound = errors.New("pricing tier not found")

// ErrDefaultPricingTierMissing error
var ErrDefaultPricingTierMissing = errors.New("pricing tier missing")

// ProductPricingRow represents a row in the product_pricing table
type ProductPricingRow struct {
	id            int
	UUID          string
	pricingTierID int
	productID     int
	UnitPrice     int
	Created       time.Time
	Modified      time.Time
}

// ProductPricingJoinRow represents a 3-way join between pricing_products,
// products and pricing_tier.
type ProductPricingJoinRow struct {
	id              int
	UUID            string
	productID       int
	ProductUUID     string
	SKU             string
	pricingTierID   *int
	PricingTierUUID string
	TierRef         string
	UnitPrice       int
	Created         time.Time
	Modified        time.Time
}

// GetProductPricing returns a ProductPricing for a given product and pricing tier.
func (m *PgModel) GetProductPricing(ctx context.Context, productUUID, tierUUID string) (*ProductPricingJoinRow, error) {
	query := `
		SELECT
                  R.id, R.uuid, P.uuid AS product_uuid, T.uuid AS pricing_tier_uuid, unit_price, R.created, R.modified
                FROM
                  product_pricing R
		JOIN product P
		  ON P.id = R.product_id
		JOIN pricing_tier T
		  ON T.id = R.pricing_tier_id
                WHERE
                  P.uuid = $1 AND T.uuid = $2;
	`
	p := ProductPricingJoinRow{}
	if err := m.db.QueryRowContext(ctx, query, productUUID, tierUUID).Scan(&p.id, &p.UUID, &p.productID, &p.ProductUUID, &p.SKU, &p.pricingTierID, &p.PricingTierUUID, &p.TierRef, &p.UnitPrice, &p.Created, &p.Modified); err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrProductPricingNotFound
		}
		return nil, errors.Wrapf(err, " m.db.QueryRowContext(ctx, %q, productUUID=%q, tierUUID=%q).Scan(...)", query, productUUID, tierUUID)
	}
	return &p, nil
}

// GetProductPricingByProductUUID returns a list of ProductPricings for a given product.
func (m *PgModel) GetProductPricingByProductUUID(ctx context.Context, productUUID string) ([]*ProductPricingJoinRow, error) {
	query := `
		SELECT
		  r.id, r.uuid AS uuid, p.id AS product_id, p.uuid as product_uuid, p.sku,
		  t.id as pricing_tier_id, t.uuid as pricing_tier_uuid, t.tier_ref, r.unit_price, r.created, r.modified
		FROM product AS p
		LEFT OUTER JOIN product_pricing AS r
		  ON p.id = r.product_id
		LEFT OUTER JOIN pricing_tier AS t
		  ON t.id = r.pricing_tier_id
		WHERE p.uuid = $1
		ORDER BY t.tier_ref
	`
	rows, err := m.db.QueryContext(ctx, query, productUUID)
	if err != nil {
		return nil, errors.Wrapf(err, "m.db.QueryContext(ctx, query=%q, productUUID=%q)", query, productUUID)
	}
	defer rows.Close()
	pricing := make([]*ProductPricingJoinRow, 0, 8)
	for rows.Next() {
		var p ProductPricingJoinRow
		if err = rows.Scan(&p.id, &p.UUID, &p.productID, &p.ProductUUID, &p.SKU,
			&p.pricingTierID, &p.PricingTierUUID, &p.TierRef, &p.UnitPrice, &p.Created, &p.Modified); err != nil {
			if err == sql.ErrNoRows {
				return nil, ErrProductPricingNotFound
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
		if pricing[0].pricingTierID == nil {
			return nil, nil
		}
	}
	return pricing, nil
}

// GetProductPricingBySKU returns a list of ProductPricing items for a given SKU.
func (m *PgModel) GetProductPricingBySKU(ctx context.Context, sku string) ([]*ProductPricingRow, error) {
	query := `
		SELECT id, pricing_tier_id, product_id, unit_price, created, modified
		FROM product_pricing
		WHERE sku = $1
		ORDER BY tier_ref ASC
	`
	rows, err := m.db.QueryContext(ctx, query, sku)
	if err != nil {
		return nil, errors.Wrapf(err, "m.db.QueryContext(ctx, %q, %q)", query, sku)
	}
	defer rows.Close()
	pricing := make([]*ProductPricingRow, 0, 8)
	for rows.Next() {
		var p ProductPricingRow
		if err = rows.Scan(&p.id, &p.pricingTierID, &p.productID, &p.UnitPrice, &p.Created, &p.Modified); err != nil {
			return nil, errors.Wrap(err, "scan failed")
		}
		pricing = append(pricing, &p)
	}
	if err = rows.Err(); err != nil {
		return nil, errors.Wrap(err, "rows.Err()")
	}
	return pricing, nil
}

// GetProductPricingID returns a list of ProductPricing items for a given tier id.
func (m *PgModel) GetProductPricingID(ctx context.Context, pricingTierID string) ([]*ProductPricingRow, error) {
	query := `
		SELECT
		  id, uuid, pricing_tier_id, product_id, unit_price, created, modified
		FROM product_pricing
		WHERE uuid = $1
	`
	rows, err := m.db.QueryContext(ctx, query, pricingTierID)
	if err != nil {
		return nil, errors.Wrapf(err, "m.db.QueryContext(ctx, %q, %q)", query, pricingTierID)
	}
	defer rows.Close()
	pricing := make([]*ProductPricingRow, 0, 8)
	for rows.Next() {
		var p ProductPricingRow
		if err = rows.Scan(&p.id, &p.pricingTierID, &p.productID, &p.UnitPrice, &p.Created, &p.Modified); err != nil {
			return nil, errors.Wrap(err, "scan failed")
		}
		pricing = append(pricing, &p)
	}
	if err = rows.Err(); err != nil {
		return nil, errors.Wrap(err, "rows.Err()")
	}
	return pricing, nil
}

// UpdateTierPricing updates the product pricing with the new unit price
// by product uuid and pricing tier uuid.
func (m *PgModel) UpdateTierPricing(ctx context.Context, productUUID, pricingTierUUID string, unitPrice float64) (*ProductPricingJoinRow, error) {
	tx, err := m.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, errors.Wrap(err, "db.BeginTx")
	}

	query := `SELECT id FROM product WHERE uuid = $1`
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

	query = `SELECT id FROM pricing_tier WHERE uuid = $1`
	var pricingTierID int
	err = tx.QueryRowContext(ctx, query, pricingTierUUID).Scan(&pricingTierID)
	if err != nil {
		if err == sql.ErrNoRows {
			tx.Rollback()
			return nil, ErrPricingTierNotFound
		}
		tx.Rollback()
		return nil, errors.Wrapf(err, "query row context failed for query=%q", query)
	}

	query = `
		UPDATE product_pricing
		SET
		  unit_price = $1, modified = NOW()
		WHERE
		  product_id = $2 AND pricing_tier_id = $3
		RETURNING
		  id, uuid, pricing_tier_id, tier_ref as (select tier_ref FROM pricing_tiers WHERE id = $4) product_id, unit_price, created, modified
	`
	p := ProductPricingJoinRow{}
	if err := tx.QueryRowContext(ctx, query, unitPrice, productID, pricingTierID, pricingTierID).Scan(&p.id, &p.UUID, &p.pricingTierID, &p.productID, &p.UnitPrice, &p.Created, &p.Modified); err != nil {
		return nil, errors.Wrap(err, "UpdateTierPricing QueryRowContext failed")
	}

	if err = tx.Commit(); err != nil {
		return nil, errors.Wrap(err, "tx.Commit")
	}
	return &p, nil
}

// DeleteProductPricing deletes a tier pricing by product uuid and pricing tier uuid.
func (m *PgModel) DeleteProductPricing(ctx context.Context, productUUID, pricingTierUUID string) error {
	tx, err := m.db.BeginTx(ctx, nil)
	if err != nil {
		return errors.Wrap(err, "db.BeginTx")
	}

	query := `SELECT id FROM product WHERE uuid = $1`
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

	query = `SELECT id FROM pricing_tier WHERE uuid = $1`
	var pricingTierID int
	err = tx.QueryRowContext(ctx, query, pricingTierUUID).Scan(&pricingTierID)
	if err != nil {
		if err == sql.ErrNoRows {
			tx.Rollback()
			return ErrPricingTierNotFound
		}
		tx.Rollback()
		return errors.Wrapf(err, "query row context failed for query=%q", query)
	}

	query = `
		DELETE FROM product_pricing
		WHERE
		  product_id = $1 AND pricing_tier_id = $2
	`
	if _, err := m.db.ExecContext(ctx, query, productID, pricingTierID); err != nil {
		tx.Rollback()
		return errors.Wrapf(err, "exec context query=%q", query)
	}

	if err = tx.Commit(); err != nil {
		return errors.Wrap(err, "tx.Commit")
	}
	return nil
}
