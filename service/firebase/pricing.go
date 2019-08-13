package firebase

import (
	"context"
	"time"

	"github.com/pkg/errors"
)

// ErrDefaultPricingTierMissing error
var ErrDefaultPricingTierMissing = errors.New("default pricing tier missing")

// PricingTierID for a pricing tier id.
type PricingTierID string

// ProductPricing contains pricing information for a single SKU and tier ref.
type ProductPricing struct {
	ProductID     string        `json:"product_id,omitempty"`
	PricingTierID PricingTierID `json:"pricing_tier_id,omitempty"`
	UnitPrice     int           `json:"unit_price"`
	Created       time.Time     `json:"created"`
	Modified      time.Time     `json:"modified"`
}

// PricingEntry represents a single pricing entry
type PricingEntry struct {
	UnitPrice int       `json:"unit_price"`
	Created   time.Time `json:"created"`
	Modified  time.Time `json:"modified"`
}

// GetTierPricing returns a ProductPricing for the product with the
// given SKU and tier ref.
func (s *Service) GetTierPricing(ctx context.Context, productID, pricingTierID string) (*ProductPricing, error) {
	p, err := s.model.GetProductPricing(ctx, productID, pricingTierID)
	if err != nil {
		return nil, errors.Wrap(err, "GetProductPricingBySKUAndTier failed")
	}
	pricing := ProductPricing{
		PricingTierID: PricingTierID(p.PricingTierUUID),
		ProductID:     p.ProductUUID,
		UnitPrice:     p.UnitPrice,
		Created:       p.Created,
		Modified:      p.Modified,
	}
	return &pricing, nil
}

// ProductTierPricing contains pricing information for all tiers of
// a given SKU.
type ProductTierPricing struct {
	UnitPrice float64 `json:"unit_price"`
}

// PricingMapByProductID returns a map of pricing tier id to PricingEntrys.
func (s *Service) PricingMapByProductID(ctx context.Context, productID string) (map[PricingTierID]*PricingEntry, error) {
	plist, err := s.model.GetProductPricingByProductUUID(ctx, productID)
	if err != nil {
		return nil, errors.Wrap(err, "GetProductPricingByProductUUID failed")
	}
	pmap := make(map[PricingTierID]*PricingEntry)
	for _, p := range plist {
		ptp := PricingEntry{
			UnitPrice: p.UnitPrice,
			Created:   p.Created,
			Modified:  p.Modified,
		}
		if _, ok := pmap[PricingTierID(p.PricingTierUUID)]; !ok {
			pmap[PricingTierID(p.TierRef)] = &ptp
		}
	}
	return pmap, nil
}

// PricingMapByTier returns a map of SKU to PricingEntrys.
func (s *Service) PricingMapByTier(ctx context.Context, ref string) (map[string]*PricingEntry, error) {
	plist, err := s.model.GetProductPricingID(ctx, ref)
	if err != nil {
		return nil, errors.Wrap(err, "GetProductPricingByTier failed")
	}
	pmap := make(map[string]*PricingEntry)
	for _, p := range plist {
		ptp := PricingEntry{
			UnitPrice: p.UnitPrice,
			Created:   p.Created,
			Modified:  p.Modified,
		}
		if _, ok := pmap[p.UUID]; !ok {
			pmap[p.UUID] = &ptp
		}
	}
	return pmap, nil
}

// ErrPricingTierNotFound is where no tier pricing could be found
// for the given sku and tier ref.
var ErrPricingTierNotFound = errors.New("service: tier pricing not found")

// UpdateTierPricing updates the tier pricing for the given sku and tier ref.
// If the produt pricing is not found returns nil, nil.
func (s *Service) UpdateTierPricing(ctx context.Context, productID, pricingTierID string, unitPrice float64) (*ProductPricing, error) {
	p, err := s.model.GetProductPricing(ctx, productID, pricingTierID)
	if err != nil {
		return nil, errors.Wrapf(err, "GetProductPricingBySKUAndTier(ctx, %q, %q) failed", productID, pricingTierID)
	}
	if p == nil {
		return nil, ErrPricingTierNotFound
	}
	p, err = s.model.UpdateTierPricing(ctx, productID, pricingTierID, unitPrice)
	if err != nil {
		return nil, errors.Wrapf(err, "UpdateTierPricing(ctx, %q, %q, %.4f) failed", productID, pricingTierID, unitPrice)
	}
	pricing := ProductPricing{
		ProductID:     p.ProductUUID,
		PricingTierID: PricingTierID(p.PricingTierUUID),
		UnitPrice:     p.UnitPrice,
		Created:       p.Created,
		Modified:      p.Modified,
	}
	return &pricing, nil
}

// DeleteTierPricing deletes a tier pricing by SKU and tier ref.
func (s *Service) DeleteTierPricing(ctx context.Context, productID, pricingTierID string) error {
	if err := s.model.DeleteProductPricing(ctx, productID, pricingTierID); err != nil {
		return errors.Wrapf(err, "DeleteProductPricingBySKUAndTier(ctx, %q, %q) failed", productID, pricingTierID)
	}
	return nil
}
