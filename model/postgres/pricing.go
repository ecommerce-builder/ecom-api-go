package postgres

import (
	"context"
	"database/sql"
	"time"

	"github.com/pkg/errors"
)

// ProductPricing maps a product_pricing table row
type ProductPricing struct {
	id        int
	TierRef   string
	SKU       string
	UnitPrice float64
	Created   time.Time
	Modified  time.Time
}

// GetProductPricingBySKUAndTier returns a ProductPricing for a given SKU and tier ref.
func (m *PgModel) GetProductPricingBySKUAndTier(ctx context.Context, sku, ref string) (*ProductPricing, error) {
	query := `
		SELECT id, tier_ref, sku, unit_price, created, modified
		FROM product_pricing
		WHERE sku = $1 AND tier_ref = $2
	`
	p := ProductPricing{}
	if err := m.db.QueryRowContext(ctx, query, sku, ref).Scan(&p.id, &p.TierRef, &p.SKU, &p.UnitPrice, &p.Created, &p.Modified); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, errors.Wrapf(err, " m.db.QueryRowContext(ctx, %q, %q, %q).Scan(...)", query, sku, ref)
	}
	return &p, nil
}

// GetProductPricingBySKU returns a list of ProductPricing items for a given SKU.
func (m *PgModel) GetProductPricingBySKU(ctx context.Context, sku string) ([]*ProductPricing, error) {
	query := `
	SELECT id, tier_ref, sku, unit_price, created, modified
	FROM product_pricing
	WHERE sku = $1
	ORDER BY tier_ref ASC
	`
	rows, err := m.db.QueryContext(ctx, query, sku)
	if err != nil {
		return nil, errors.Wrapf(err, "m.db.QueryContext(ctx, %q, %q)", query, sku)
	}
	defer rows.Close()
	pricing := make([]*ProductPricing, 0, 8)
	for rows.Next() {
		var p ProductPricing
		if err = rows.Scan(&p.id, &p.TierRef, &p.SKU, &p.UnitPrice, &p.Created, &p.Modified); err != nil {
			return nil, errors.Wrap(err, "scan failed")
		}
		pricing = append(pricing, &p)
	}
	if err = rows.Err(); err != nil {
		return nil, errors.Wrap(err, "rows.Err()")
	}
	return pricing, nil
}

// GetProductPricingByTier returns a list of ProductPricing items for a given tier ref.
func (m *PgModel) GetProductPricingByTier(ctx context.Context, ref string) ([]*ProductPricing, error) {
	query := `
	SELECT id, tier_ref, sku, unit_price, created, modified
	FROM product_pricing
	WHERE tier_ref = $1
	`
	rows, err := m.db.QueryContext(ctx, query, ref)
	if err != nil {
		return nil, errors.Wrapf(err, "m.db.QueryContext(ctx, %q, %q)", query, ref)
	}
	defer rows.Close()
	pricing := make([]*ProductPricing, 0, 8)
	for rows.Next() {
		var p ProductPricing
		if err = rows.Scan(&p.id, &p.TierRef, &p.SKU, &p.UnitPrice, &p.Created, &p.Modified); err != nil {
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
// by `sku` and tier `ref`.
func (m *PgModel) UpdateTierPricing(ctx context.Context, sku, ref string, unitPrice float64) (*ProductPricing, error) {
	query := `
		UPDATE product_pricing
		SET
		  unit_price = $1, modified = NOW()
		WHERE
		  sku = $2 AND tier_ref = $3
		RETURNING
		  id, tier_ref, sku, unit_price, created, modified
	`
	p := ProductPricing{}
	if err := m.db.QueryRowContext(ctx, query, unitPrice, sku, ref).Scan(&p.id, &p.TierRef, &p.SKU, &p.UnitPrice, &p.Created, &p.Modified); err != nil {
		return nil, errors.Wrap(err, "UpdateTierPricing QueryRowContext failed")
	}
	return &p, nil
}

// DeleteProductPricingBySKUAndTier deletes a tier pricing by SKU and tier ref.
func (m *PgModel) DeleteProductPricingBySKUAndTier(ctx context.Context, sku, ref string) error {
	query := `DELETE FROM product_pricing WHERE sku = $1 AND tier_ref = $2`
	if _, err := m.db.ExecContext(ctx, query, sku, ref); err != nil {
		return errors.Wrapf(err, "exec context query=%q", query)
	}
	return nil
}
