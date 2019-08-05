package firebase

import (
	"context"
	"time"

	"bitbucket.org/andyfusniakteam/ecom-api-go/model/postgres"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

var (
	// ErrCartNotFound is returned when attempting an operation on non existing cart.
	ErrCartNotFound = errors.New("cart not found")

	// ErrCartItemAlreadyExists is returned when attempting to add a
	// cart item to a cart that already contains that item.
	ErrCartItemAlreadyExists = errors.New("cart already exists")

	// ErrCartItemNotFound is returned when a cart ID cannot be found.
	ErrCartItemNotFound = errors.New("cart item not found")

	// ErrCartContainsNoItems occurs when attempting to delete all items.
	ErrCartContainsNoItems = errors.New("cart contains no items")
)

// Cart holds the details of a shopping cart.
type Cart struct {
	ID       string    `json:"id"`
	Locked   bool      `json:"locked"`
	Created  time.Time `json:"created"`
	Modified time.Time `json:"modified"`
}

// CartItem structure holds the details individual cart item
type CartItem struct {
	SKU       string    `json:"sku"`
	Name      string    `json:"name"`
	Qty       int       `json:"qty"`
	UnitPrice int       `json:"unit_price"`
	Created   time.Time `json:"created"`
	Modified  time.Time `json:"modified"`
}

// CreateCart generates a new random id to be used for subseqent cart calls.
func (s *Service) CreateCart(ctx context.Context) (*Cart, error) {
	cartRow, err := s.model.CreateCart(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "service: s.model.CreateCart(ctx) failed")
	}
	cart := Cart{
		ID:       cartRow.UUID,
		Locked:   cartRow.Locked,
		Created:  cartRow.Created,
		Modified: cartRow.Modified,
	}
	return &cart, nil
}

// AddItemToCart adds a single item to a given cart.
// Returns `ErrCartNotFound` if the cart with `cartID` does not exist.
func (s *Service) AddItemToCart(ctx context.Context, cartID string, sku string, qty int) (*CartItem, error) {
	log.WithContext(ctx).Debugf("service: s.AddItemToCart(cartID=%q, sku=%q, qty=%d) started", cartID, sku, qty)

	exists, _ := s.model.IsCartExists(ctx, cartID)
	if !exists {
		return nil, ErrCartNotFound
	}

	item, err := s.model.AddItemToCart(ctx, cartID, "default", sku, qty)
	if err != nil {
		if err == postgres.ErrCartItemAlreadyExists {
			return nil, ErrCartItemAlreadyExists
		}
		return nil, errors.Wrapf(err, "s.model.AddItemToCart(ctx, %q, %q, %q, %d) failed: ", cartID, "default", sku, qty)
	}
	sitem := CartItem{
		SKU:       item.SKU,
		Name:      item.Name,
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

// GetCartItems get all cart items by cart ID.
func (s *Service) GetCartItems(ctx context.Context, cartID string) ([]*CartItem, error) {
	items, err := s.model.GetCartItems(ctx, cartID)
	if err != nil {
		if err == postgres.ErrCartNotFound {
			return nil, ErrCartNotFound
		}
		return nil, err
	}

	results := make([]*CartItem, 0, 32)
	for _, v := range items {
		i := CartItem{
			SKU:       v.SKU,
			Name:      v.Name,
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
func (s *Service) UpdateCartItem(ctx context.Context, cartID, sku string, qty int) (*CartItem, error) {
	item, err := s.model.UpdateItemByCartUUID(ctx, cartID, sku, qty)
	if err != nil {
		if err == postgres.ErrCartItemNotFound {
			return nil, ErrCartItemNotFound
		}
		return nil, err
	}
	sitem := CartItem{
		SKU:       item.SKU,
		Name:      item.Name,
		Qty:       item.Qty,
		UnitPrice: item.UnitPrice,
		Created:   item.Created,
		Modified:  item.Modified,
	}
	return &sitem, nil
}

// DeleteCartItem deletes a single cart item
func (s *Service) DeleteCartItem(ctx context.Context, cartID, sku string) (count int64, err error) {
	count, err = s.model.DeleteCartItem(ctx, cartID, sku)
	if err != nil {
		if err == postgres.ErrCartNotFound {
			return 0, ErrCartNotFound
		}
		return -1, err
	}
	return count, nil
}

// EmptyCartItems empties the cart of all items but not coupons
func (s *Service) EmptyCartItems(ctx context.Context, cartID string) error {
	err := s.model.EmptyCartItems(ctx, cartID)
	if err != nil {
		if err == postgres.ErrCartNotFound {
			return ErrCartNotFound
		} else if err == postgres.ErrCartContainsNoItems {
			return ErrCartContainsNoItems
		}
		return err
	}
	return nil
}
