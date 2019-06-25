package firebase

import (
	"context"
	"time"

	"github.com/pkg/errors"
)

// SKU for a product
type SKU string

// TierRef for a pricing tier reference.
type TierRef string

// ProductPricing contains pricing information for a single SKU and tier ref.
type ProductPricing struct {
	SKU       SKU       `json:"sku,omitempty"`
	TierRef   TierRef   `json:"tier_ref,omitempty"`
	UnitPrice float64   `json:"unit_price"`
	Created   time.Time `json:"created"`
	Modified  time.Time `json:"modified"`
}

// PricingEntry represents a single pricing entry
type PricingEntry struct {
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
		TierRef:   TierRef(p.TierRef),
		SKU:       SKU(p.SKU),
		UnitPrice: p.UnitPrice,
		Created:   p.Created,
		Modified:  p.Modified,
	}
	return &pricing, nil
}

// ProductTierPricing contains pricing information for all tiers of
// a given SKU.
type ProductTierPricing struct {
	UnitPrice float64 `json:"unit_price"`
}

// PricingMapBySKU returns a map of tier to PricingEntrys.
func (s *Service) PricingMapBySKU(ctx context.Context, sku string) (map[TierRef]*PricingEntry, error) {
	plist, err := s.model.GetProductPricingBySKU(ctx, sku)
	if err != nil {
		return nil, errors.Wrap(err, "GetProductPricingBySKU failed")
	}
	pmap := make(map[TierRef]*PricingEntry)
	for _, p := range plist {
		ptp := PricingEntry{
			UnitPrice: p.UnitPrice,
			Created:   p.Created,
			Modified:  p.Modified,
		}
		if _, ok := pmap[TierRef(p.TierRef)]; !ok {
			pmap[TierRef(p.TierRef)] = &ptp
		}
	}
	return pmap, nil
}

// PricingMapByTier returns a map of SKU to PricingEntrys.
func (s *Service) PricingMapByTier(ctx context.Context, ref string) (map[SKU]*PricingEntry, error) {
	plist, err := s.model.GetProductPricingByTier(ctx, ref)
	if err != nil {
		return nil, errors.Wrap(err, "GetProductPricingByTier failed")
	}
	pmap := make(map[SKU]*PricingEntry)
	for _, p := range plist {
		ptp := PricingEntry{
			UnitPrice: p.UnitPrice,
			Created:   p.Created,
			Modified:  p.Modified,
		}
		if _, ok := pmap[SKU(p.SKU)]; !ok {
			pmap[SKU(p.SKU)] = &ptp
		}
	}
	return pmap, nil
}

// ErrTierPricingNotFound is where no tier pricing could be found
// for the given sku and tier ref.
var ErrTierPricingNotFound = errors.New("service: tier pricing not found")

// UpdateTierPricing updates the tier pricing for the given sku and tier ref.
// If the produt pricing is not found returns nil, nil.
func (s *Service) UpdateTierPricing(ctx context.Context, sku, tierRef string, unitPrice float64) (*ProductPricing, error) {
	p, err := s.model.GetProductPricingBySKUAndTier(ctx, sku, tierRef)
	if err != nil {
		return nil, errors.Wrapf(err, "GetProductPricingBySKUAndTier(ctx, %q, %q) failed", sku, tierRef)
	}
	if p == nil {
		return nil, ErrTierPricingNotFound
	}
	p, err = s.model.UpdateTierPricing(ctx, sku, tierRef, unitPrice)
	if err != nil {
		return nil, errors.Wrapf(err, "UpdateTierPricing(ctx, %q, %q, %.4f) failed", sku, tierRef, unitPrice)
	}
	pricing := ProductPricing{
		SKU:       SKU(p.SKU),
		TierRef:   TierRef(p.TierRef),
		UnitPrice: p.UnitPrice,
		Created:   p.Created,
		Modified:  p.Modified,
	}
	return &pricing, nil
}

// DeleteTierPricing deletes a tier pricing by SKU and tier ref.
func (s *Service) DeleteTierPricing(ctx context.Context, sku, tierRef string) error {
	if err := s.model.DeleteProductPricingBySKUAndTier(ctx, sku, tierRef); err != nil {
		return errors.Wrapf(err, "DeleteProductPricingBySKUAndTier(ctx, %q, %q) failed", sku, tierRef)
	}
	return nil
}
