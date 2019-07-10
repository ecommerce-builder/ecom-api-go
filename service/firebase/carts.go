package firebase

import (
	"context"
	"time"

	"bitbucket.org/andyfusniakteam/ecom-api-go/model/postgres"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

var (
	// ErrCartItemAlreadyExists is returned when attempting to add a
	// cart item to a cart that already contains that item.
	ErrCartItemAlreadyExists = errors.New("cart already exists")

	ErrCartItemNotFound = errors.New("cart item not found")
)

// CartItem structure holds the details individual cart item
type CartItem struct {
	SKU       string    `json:"sku"`
	Qty       int       `json:"qty"`
	UnitPrice int       `json:"unit_price"`
	Created   time.Time `json:"created"`
	Modified  time.Time `json:"modified"`
}

// CreateCart generates a new random id to be used for subseqent cart calls
func (s *Service) CreateCart(ctx context.Context) (*string, error) {
	strptr, err := s.model.CreateCart(ctx)
	if err != nil {
		return nil, err
	}
	return strptr, nil
}

// AddItemToCart adds a single item to a given cart
func (s *Service) AddItemToCart(ctx context.Context, id string, sku string, qty int) (*CartItem, error) {
	log.WithContext(ctx).Debugf("s.AddItemToCart(%q, %q, %d) started", id, sku, qty)
	item, err := s.model.AddItemToCart(ctx, id, "default", sku, qty)
	if err != nil {
		if err == postgres.ErrCartItemAlreadyExists {
			return nil, ErrCartItemAlreadyExists
		}
		return nil, errors.Wrapf(err, "s.model.AddItemToCart(ctx, %q, %q, %q, %d) failed: ", id, "default", sku, qty)
	}
	sitem := CartItem{
		SKU:       item.SKU,
		Qty:       item.Qty,
		UnitPrice: item.UnitPrice,
		Created:   item.Created,
		Modified:  item.Modified,
	}
	return &sitem, nil
}

// HasCartItems returns true if any cart items have previously been added to the given cart.
func (s *Service) HasCartItems(ctx context.Context, id string) (bool, error) {
	has, err := s.model.HasCartItems(ctx, id)
	if err != nil {
		return false, errors.Wrap(err, "service: has cart items failed")
	}
	return has, nil
}

// GetCartItems get all items in the given cart
func (s *Service) GetCartItems(ctx context.Context, id string) ([]*CartItem, error) {
	items, err := s.model.GetCartItems(ctx, id)
	if err != nil {
		return nil, err
	}

	results := make([]*CartItem, 0, 32)
	for _, v := range items {
		i := CartItem{
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
func (s *Service) UpdateCartItem(ctx context.Context, id string, sku string, qty int) (*CartItem, error) {
	item, err := s.model.UpdateItemByCartID(ctx, id, sku, qty)
	if err != nil {
		if err == postgres.ErrCartItemNotFound {
			return nil, ErrCartItemNotFound
		}
		return nil, err
	}
	sitem := CartItem{
		SKU:       item.SKU,
		Qty:       item.Qty,
		UnitPrice: item.UnitPrice,
		Created:   item.Created,
		Modified:  item.Modified,
	}
	return &sitem, nil
}

// DeleteCartItem deletes a single cart item
func (s *Service) DeleteCartItem(ctx context.Context, id string, sku string) (count int64, err error) {
	count, err = s.model.DeleteCartItem(ctx, id, sku)
	if err != nil {
		return -1, err
	}
	return count, nil
}

// EmptyCartItems empties the cart of all items but not coupons
func (s *Service) EmptyCartItems(ctx context.Context, id string) (err error) {
	return s.model.EmptyCartItems(ctx, id)
}
