package firebase

import (
	"context"
	"time"

	log "github.com/sirupsen/logrus"
)

// CartItem structure holds the details individual cart item
type CartItem struct {
	UUID      string    `json:"uuid"`
	SKU       string    `json:"sku"`
	Qty       int       `json:"qty"`
	UnitPrice float64   `json:"unit_price"`
	Created   time.Time `json:"created"`
	Modified  time.Time `json:"modified"`
}

// CreateCart generates a new random UUID to be used for subseqent cart calls
func (s *Service) CreateCart(ctx context.Context) (*string, error) {
	strptr, err := s.model.CreateCart(ctx)
	if err != nil {
		return nil, err
	}
	return strptr, nil
}

// AddItemToCart adds a single item to a given cart
func (s *Service) AddItemToCart(ctx context.Context, uuid string, sku string, qty int) (*CartItem, error) {
	log.Debugf("s.AddItemToCart(%s, %s, %d) started", uuid, sku, qty)
	item, err := s.model.AddItemToCart(ctx, uuid, "default", sku, qty)
	if err != nil {
		return nil, err
	}
	sitem := CartItem{
		UUID:      item.UUID,
		SKU:       item.SKU,
		Qty:       item.Qty,
		UnitPrice: item.UnitPrice,
		Created:   item.Created,
		Modified:  item.Modified,
	}
	return &sitem, nil
}

// GetCartItems get all items in the given cart
func (s *Service) GetCartItems(ctx context.Context, cartUUID string) ([]*CartItem, error) {
	items, err := s.model.GetCartItems(ctx, cartUUID)
	if err != nil {
		return nil, err
	}

	results := make([]*CartItem, 0, 32)
	for _, v := range items {
		i := CartItem{
			UUID:      v.UUID,
			SKU:       v.SKU,
			Qty:       v.Qty,
			UnitPrice: v.UnitPrice,
			Created:   v.Created,
			Modified:  v.Modified,
		}
		results = append(results, &i)
	}
	return results, nil
}

// UpdateCartItem updates a single item's qty
func (s *Service) UpdateCartItem(ctx context.Context, cartUUID string, sku string, qty int) (*CartItem, error) {
	item, err := s.model.UpdateItemByCartUUID(ctx, cartUUID, sku, qty)
	if err != nil {
		return nil, err
	}
	sitem := CartItem{
		UUID:      item.UUID,
		SKU:       item.SKU,
		Qty:       item.Qty,
		UnitPrice: item.UnitPrice,
		Created:   item.Created,
		Modified:  item.Modified,
	}
	return &sitem, nil
}

// DeleteCartItem deletes a single cart item
func (s *Service) DeleteCartItem(ctx context.Context, cartUUID string, sku string) (count int64, err error) {
	count, err = s.model.DeleteCartItem(ctx, cartUUID, sku)
	if err != nil {
		return -1, err
	}
	return count, nil
}

// EmptyCartItems empties the cart of all items but not coupons
func (s *Service) EmptyCartItems(ctx context.Context, cartUUID string) (err error) {
	return s.model.EmptyCartItems(ctx, cartUUID)
}
