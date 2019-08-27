package firebase

import (
	"context"
	"fmt"
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
	Object    string    `json:"object"`
	ID        string    `json:"id"`
	CartID    string    `json:"cart_id,omitempty"`
	ProductID string    `json:"product_id"`
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
func (s *Service) AddItemToCart(ctx context.Context, cartID, productID string, qty int) (*CartItem, error) {
	userID := ctx.Value("cid").(string)
	fmt.Printf("UserID = %#v\n", userID)

	log.WithContext(ctx).Debugf("service: s.AddItemToCart(cartID=%q, userID=%q, productID=%q, qty=%d) started", cartID, userID, productID, qty)

	item, err := s.model.AddItemToCart(ctx, cartID, userID, productID, qty)
	if err != nil {
		if err == postgres.ErrCartNotFound {
			return nil, ErrCartNotFound
		} else if err == postgres.ErrUserNotFound {
			return nil, ErrUserNotFound
		} else if err == postgres.ErrProductNotFound {
			return nil, ErrProductNotFound
		} else if err == postgres.ErrDefaultPriceListMissing {
			return nil, ErrDefaultPriceListMissing
		} else if err == postgres.ErrCartItemAlreadyExists {
			return nil, ErrCartItemAlreadyExists
		}
		return nil, errors.Wrapf(err, "s.model.AddItemToCart(ctx, cartID=%q, %q, productID=%q, qty=%d) failed: ", cartID, "default", productID, qty)
	}
	sitem := CartItem{
		Object:    "cart_item",
		ID:        item.UUID,
		ProductID: item.ProductUUID,
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
	userUUID := ctx.Value("cid").(string)

	items, err := s.model.GetCartItems(ctx, cartID, userUUID)
	if err != nil {
		if err == postgres.ErrCartNotFound {
			return nil, ErrCartNotFound
		}
		return nil, err
	}

	results := make([]*CartItem, 0, 32)
	for _, v := range items {
		i := CartItem{
			Object:    "cart_item",
			ID:        v.UUID,
			ProductID: v.ProductUUID,
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

// UpdateCartItem updates a cart item quantity.
func (s *Service) UpdateCartItem(ctx context.Context, cartID, productID string, qty int) (*CartItem, error) {
	userUUID := ctx.Value("cid").(string)

	item, err := s.model.UpdateItemByCartUUID(ctx, cartID, userUUID, productID, qty)
	if err != nil {
		if err == postgres.ErrCartNotFound {
			return nil, ErrCartNotFound
		} else if err == postgres.ErrProductNotFound {
			return nil, ErrProductNotFound
		} else if err == postgres.ErrCartItemNotFound {
			return nil, ErrCartItemNotFound
		}
		return nil, err
	}
	sitem := CartItem{
		Object:    "cart_item",
		ID:        item.UUID,
		ProductID: item.ProductUUID,
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
func (s *Service) DeleteCartItem(ctx context.Context, cartID, productID string) error {
	if err := s.model.DeleteCartItem(ctx, cartID, productID); err != nil {
		if err == postgres.ErrCartNotFound {
			return ErrCartNotFound
		} else if err == postgres.ErrProductNotFound {
			return ErrProductNotFound
		} else if err == postgres.ErrCartItemNotFound {
			return ErrCartItemNotFound
		}
		return errors.Wrapf(err, "s.model.DeleteCartItem(ctx, cartID=%q, productID=%q) failed", cartID, productID)
	}
	return nil
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
