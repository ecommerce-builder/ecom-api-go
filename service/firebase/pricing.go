package firebase

import (
	"context"
	"fmt"
	"time"

	"github.com/pkg/errors"
)

// ProductPricing contains pricing information for a single SKU and tier ref.
type ProductPricing struct {
	SKU       string    `json:"sku"`
	TierRef   string    `json:"tier_ref"`
	UnitPrice float64   `json:"unit_price"`
	Created   time.Time `json:"created"`
	Modified  time.Time `json:"modified"`
}

// GetTierPricing returns a ProductPricing for the product with the
// given SKU and tier ref.
func (s *Service) GetTierPricing(ctx context.Context, sku, ref string) (*ProductPricing, error) {
	p, err := s.model.GetProductPricingBySKUAndTier(ctx, sku, ref)
	if err != nil {
		return nil, errors.Wrap(err, "GetProductPricingBySKUAndTier failed")
	}
	pricing := ProductPricing{
		TierRef:   p.TierRef,
		SKU:       p.SKU,
		UnitPrice: p.UnitPrice,
		Created:   p.Created,
		Modified:  p.Modified,
	}
	fmt.Println(pricing)
	return &pricing, nil
}

// ProductTierPricing contains pricing information for all tiers of
// a given SKU.
type ProductTierPricing struct {
	UnitPrice float64 `json:"unit_price"`
}

// ListPricingBySKU returns a map of tier to ProductTierPricings.
func (s *Service) ListPricingBySKU(ctx context.Context, sku string) (map[string]*ProductTierPricing, error) {
	plist, err := s.model.GetProductPricingBySKU(ctx, sku)
	if err != nil {
		return nil, errors.Wrap(err, "GetProductPricingBySKU failed")
	}
	pmap := make(map[string]*ProductTierPricing)
	for _, p := range plist {
		ptp := ProductTierPricing{
			UnitPrice: p.UnitPrice,
		}
		if _, ok := pmap[p.TierRef]; !ok {
			pmap[p.TierRef] = &ptp
		}
	}
	return pmap, nil
}

// ProductSKUPricing contains pricing information for all SKUs of given tier.
type ProductSKUPricing struct {
	UnitPrice float64 `json:"unit_price"`
}

// ListPricingByTier returns a map of SKU to ProductSKUPricings.
func (s *Service) ListPricingByTier(ctx context.Context, ref string) (map[string]*ProductSKUPricing, error) {
	plist, err := s.model.GetProductPricingByTier(ctx, ref)
	if err != nil {
		return nil, errors.Wrap(err, "GetProductPricingByTier failed")
	}
	pmap := make(map[string]*ProductSKUPricing)
	for _, p := range plist {
		ptp := ProductSKUPricing{
			UnitPrice: p.UnitPrice,
		}
		if _, ok := pmap[p.SKU]; !ok {
			pmap[p.SKU] = &ptp
		}
	}
	return pmap, nil
}

// ErrTierPricingNotFound is where no tier pricing could be found
// for the given sku and tier ref.
var ErrTierPricingNotFound = errors.New("service: tier pricing not found")

// UpdateTierPricing updates the tier pricing for the given sku and tier ref.
// If the produt pricing is not found returns nil, nil.
func (s *Service) UpdateTierPricing(ctx context.Context, sku, ref string, unitPrice float64) (*ProductPricing, error) {
	p, err := s.model.GetProductPricingBySKUAndTier(ctx, sku, ref)
	if err != nil {
		return nil, errors.Wrapf(err, "GetProductPricingBySKUAndTier(ctx, %q, %q) failed", sku, ref)
	}
	if p == nil {
		return nil, ErrTierPricingNotFound
	}
	p, err = s.model.UpdateTierPricing(ctx, sku, ref, unitPrice)
	if err != nil {
		return nil, errors.Wrapf(err, "UpdateTierPricing(ctx, %q, %q, %.4f) failed", sku, ref, unitPrice)
	}
	pricing := ProductPricing{
		SKU:       p.SKU,
		TierRef:   p.TierRef,
		UnitPrice: p.UnitPrice,
		Created:   p.Created,
		Modified:  p.Modified,
	}
	return &pricing, nil
}

// DeleteTierPricing deletes a tier pricing by SKU and tier ref.
func (s *Service) DeleteTierPricing(ctx context.Context, sku, ref string) error {
	if err := s.model.DeleteProductPricingBySKUAndTier(ctx, sku, ref); err != nil {
		return errors.Wrapf(err, "DeleteProductPricingBySKUAndTier(ctx, %q, %q) failed", sku, ref)
	}
	return nil
}
