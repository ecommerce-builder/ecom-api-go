package firebase

import (
	"context"
	"time"

	"github.com/pkg/errors"
)

// ErrDefaultPriceListMissing error
var ErrDefaultPriceListMissing = errors.New("default pricing list missing")

// PriceListID for a pricing list id.
type PriceListID string

// Price contains pricing information for a single product and price list.
type Price struct {
	ProductID   string    `json:"product_id,omitempty"`
	PriceListID string    `json:"pricing_list_id,omitempty"`
	UnitPrice   int       `json:"unit_price"`
	Created     time.Time `json:"created"`
	Modified    time.Time `json:"modified"`
}

// Price represents a single price
// type PricingEntry struct {
// 	UnitPrice int       `json:"unit_price"`
// 	Created   time.Time `json:"created"`
// 	Modified  time.Time `json:"modified"`
// }

// GetProductPrice returns a ProductPrice for the product with the
// given product id and price list id.
func (s *Service) GetProductPrice(ctx context.Context, productID, priceListID string) (*Price, error) {
	p, err := s.model.GetPricesByPriceList(ctx, productID, priceListID)
	if err != nil {
		return nil, errors.Wrap(err, "GetProductPricingBySKUAndTier failed")
	}
	pricing := Price{
		ProductID:   p.ProductUUID,
		PriceListID: p.PriceListUUID,
		UnitPrice:   p.UnitPrice,
		Created:     p.Created,
		Modified:    p.Modified,
	}
	return &pricing, nil
}

// ProductTierPricing contains pricing information for all tiers of
// a given SKU.
type ProductTierPricing struct {
	UnitPrice float64 `json:"unit_price"`
}

// PriceMap returns a map of pricing list id to price.
func (s *Service) PriceMap(ctx context.Context, productID string) (map[PriceListID]*Price, error) {
	plist, err := s.model.GetPrices(ctx, productID)
	if err != nil {
		return nil, errors.Wrap(err, "GetProductPricingByProductUUID failed")
	}
	pmap := make(map[PriceListID]*Price)
	for _, p := range plist {
		ptp := Price{
			UnitPrice: p.UnitPrice,
			Created:   p.Created,
			Modified:  p.Modified,
		}
		if _, ok := pmap[PriceListID(p.PriceListUUID)]; !ok {
			pmap[PriceListID(p.PriceListCode)] = &ptp
		}
	}
	return pmap, nil
}

// PriceMapByPriceList returns a map of product ids to Price.
func (s *Service) PriceMapByPriceList(ctx context.Context, priceListID string) (map[string]*Price, error) {
	plist, err := s.model.GetProductPriceByPriceList(ctx, priceListID)
	if err != nil {
		return nil, errors.Wrap(err, "GetProductPricingByTier failed")
	}
	pmap := make(map[string]*Price)
	for _, p := range plist {
		ptp := Price{
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

// UpdateTierPricing updates the tier pricing for the given sku and tier ref.
// If the produt pricing is not found returns nil, nil.
func (s *Service) UpdateTierPricing(ctx context.Context, productID, priceListID string, unitPrice float64) (*Price, error) {
	p, err := s.model.GetPricesByPriceList(ctx, productID, priceListID)
	if err != nil {
		return nil, errors.Wrapf(err, "GetProductPricingBySKUAndTier(ctx, %q, %q) failed", productID, priceListID)
	}
	if p == nil {
		return nil, ErrPriceListNotFound
	}
	p, err = s.model.UpdatePrice(ctx, productID, priceListID, unitPrice)
	if err != nil {
		return nil, errors.Wrapf(err, "UpdateTierPricing(ctx, %q, %q, %.4f) failed", productID, priceListID, unitPrice)
	}
	pricing := Price{
		ProductID:   p.ProductUUID,
		PriceListID: p.PriceListUUID,
		UnitPrice:   p.UnitPrice,
		Created:     p.Created,
		Modified:    p.Modified,
	}
	return &pricing, nil
}

// DeletePrices deletes a price list by id.
func (s *Service) DeletePrices(ctx context.Context, productID, priceListID string) error {
	if err := s.model.DeleteProductPrices(ctx, productID, priceListID); err != nil {
		return errors.Wrapf(err, "DeletePrice(ctx, productID=%q, priceListID=%q) failed", productID, priceListID)
	}
	return nil
}
