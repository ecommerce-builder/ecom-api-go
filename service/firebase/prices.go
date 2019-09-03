package firebase

import (
	"context"
	"crypto/subtle"
	"time"

	"bitbucket.org/andyfusniakteam/ecom-api-go/model/postgres"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

// ErrDefaultPriceListNotFound error
var ErrDefaultPriceListNotFound = errors.New("default pricing list not found")

// PriceListID for a pricing list id.
type PriceListID string

// Price contains pricing information for a single product and price list.
type Price struct {
	Object        string    `json:"object"`
	ID            string    `json:"id"`
	ProductID     string    `json:"product_id"`
	ProductPath   string    `json:"product_path"`
	ProductSKU    string    `json:"product_sku"`
	PriceListID   string    `json:"price_list_id"`
	PriceListCode string    `json:"price_list_code"`
	Break         int       `json:"break"`
	UnitPrice     int       `json:"unit_price"`
	Created       time.Time `json:"created"`
	Modified      time.Time `json:"modified"`
}

// PriceRequest represents a single price
type PriceRequest struct {
	Break     int `json:"break"`
	UnitPrice int `json:"unit_price"`
}

// Price represents a single price
// type PricingEntry struct {
// 	UnitPrice int       `json:"unit_price"`
// 	Created   time.Time `json:"created"`
// 	Modified  time.Time `json:"modified"`
// }

// GetPricesByProductIDAndPriceListID returns prices for the product with the
// given product id and price list id.
func (s *Service) GetPricesByProductIDAndPriceListID(ctx context.Context, productID, priceListID string) (*Price, error) {
	p, err := s.model.GetPricesByPriceList(ctx, productID, priceListID)
	if err != nil {
		return nil, errors.Wrap(err, "GetProductPricingBySKUAndTier failed")
	}
	prices := Price{
		Object:        "price",
		ProductID:     p.ProductUUID,
		ProductPath:   p.ProductPath,
		ProductSKU:    p.ProductSKU,
		PriceListID:   p.PriceListUUID,
		PriceListCode: p.PriceListCode,
		UnitPrice:     p.UnitPrice,
		Created:       p.Created,
		Modified:      p.Modified,
	}
	return &prices, nil
}

// ProductTierPricing contains pricing information for all tiers of
// a given SKU.
type ProductTierPricing struct {
	UnitPrice float64 `json:"unit_price"`
}

// UserCanAccessPriceList return true if the given user has access to the price list.
func (s *Service) UserCanAccessPriceList(ctx context.Context, userID, priceListID string) (bool, error) {
	user, err := s.model.GetUserByUUID(ctx, userID)
	if err != nil {
		if err == postgres.ErrUserNotFound {
			return false, ErrUserNotFound
		}
		return false, errors.Wrapf(err, "service: s.model.GetUserByUUID(ctx, userID=%q) failed", userID)
	}
	if subtle.ConstantTimeCompare([]byte(user.PriceListUUID), []byte(priceListID)) == 1 {
		return true, nil
	}
	return false, nil
}

// GetPrices returns a list of prices.
func (s *Service) GetPrices(ctx context.Context, productID, priceListID string) ([]*Price, error) {
	contextLogger := log.WithContext(ctx)
	contextLogger.Infof("service: GetPrices(ctx context.Context, productID=%q, priceListID=%q) started", productID, priceListID)

	plist, err := s.model.GetPrices(ctx, productID, priceListID)
	if err != nil {
		if err == postgres.ErrProductNotFound {
			return nil, ErrProductNotFound
		} else if err == postgres.ErrPriceListNotFound {
			return nil, ErrPriceListNotFound
		}
		return nil, errors.Wrapf(err, "service: GetPrices(ctx, productID=%q, priceListID=%q) failed", productID, priceListID)
	}

	prices := make([]*Price, 0, len(plist))
	for _, p := range plist {
		price := Price{
			Object:        "price",
			ID:            p.UUID,
			ProductID:     p.ProductUUID,
			ProductPath:   p.ProductPath,
			ProductSKU:    p.ProductSKU,
			PriceListID:   p.PriceListUUID,
			PriceListCode: p.PriceListCode,
			Break:         p.Break,
			UnitPrice:     p.UnitPrice,
			Created:       p.Created,
			Modified:      p.Modified,
		}
		prices = append(prices, &price)
	}
	return prices, nil
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
			Object:    "price",
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

// UpdateProductPrices updates the prices for a given product and product list.
func (s *Service) UpdateProductPrices(ctx context.Context, productID, priceListID string, createPrices []*PriceRequest) ([]*Price, error) {
	contextLogger := log.WithContext(ctx)
	contextLogger.Debugf("service: UpdateProductPrices(ctx context.Context, productID=%q, priceListID=%q, ...) started", productID, priceListID)
	contextLogger.Debugf("service: received %d prices", len(createPrices))

	cps := make([]*postgres.CreatePrice, 0, len(createPrices))
	for _, p := range createPrices {
		cp := postgres.CreatePrice{
			Break:     p.Break,
			UnitPrice: p.UnitPrice,
		}
		cps = append(cps, &cp)
	}

	plist, err := s.model.UpdatePrices(ctx, productID, priceListID, cps)
	if err != nil {
		if err == postgres.ErrProductNotFound {
			return nil, ErrProductNotFound
		} else if err == postgres.ErrPriceListNotFound {
			return nil, ErrPriceListNotFound
		}
		return nil, errors.Wrapf(err, "service: UpdatePrice(ctx, productID=%q, priceListID=%q, createPrices=%v) failed", productID, priceListID, createPrices)
	}

	prices := make([]*Price, 0, len(plist))
	for _, p := range plist {
		price := Price{
			Object:        "price",
			ID:            p.UUID,
			ProductID:     p.ProductUUID,
			ProductPath:   p.ProductPath,
			ProductSKU:    p.ProductSKU,
			PriceListID:   p.PriceListUUID,
			PriceListCode: p.PriceListCode,
			Break:         p.Break,
			UnitPrice:     p.UnitPrice,
			Created:       p.Created,
			Modified:      p.Modified,
		}
		prices = append(prices, &price)
	}
	return prices, nil
}

// DeletePrices deletes a price list by id.
func (s *Service) DeletePrices(ctx context.Context, productID, priceListID string) error {
	if err := s.model.DeleteProductPrices(ctx, productID, priceListID); err != nil {
		return errors.Wrapf(err, "DeletePrice(ctx, productID=%q, priceListID=%q) failed", productID, priceListID)
	}
	return nil
}
