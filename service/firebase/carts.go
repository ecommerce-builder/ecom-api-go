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

	// ErrCartProductExists is returned when attempting to add a
	// cart item to a cart that already contains that item.
	ErrCartProductExists = errors.New("cart product already exists")

	// ErrCartProductNotFound is returned when a cart product cannot be found.
	ErrCartProductNotFound = errors.New("cart product not found")

	// ErrCartContainsNoProducts occurs when attempting to delete all items.
	ErrCartContainsNoProducts = errors.New("cart contains no products")

	// ErrProductHasNoPrices error
	ErrProductHasNoPrices = errors.New("service: product has no prices")
)

// Cart holds the details of a shopping cart.
type Cart struct {
	ID       string    `json:"id"`
	Locked   bool      `json:"locked"`
	Created  time.Time `json:"created"`
	Modified time.Time `json:"modified"`
}

// CartProduct structure holds the details individual cart product.
type CartProduct struct {
	Object    string    `json:"object"`
	ID        string    `json:"id"`
	CartID    string    `json:"cart_id"`
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

// AddProductToCart adds a single product to a given cart.
// Returns `ErrCartNotFound` if the cart with `cartID` does not exist.
func (s *Service) AddProductToCart(ctx context.Context, userID, cartID, productID string, qty int) (*CartProduct, error) {
	log.WithContext(ctx).Debugf("service: s.AddProductToCart(userID=%q, cartID=%q, userID=%q, productID=%q, qty=%d) started", userID, cartID, userID, productID, qty)

	item, err := s.model.AddProductToCart(ctx, cartID, userID, productID, qty)
	if err != nil {
		if err == postgres.ErrCartNotFound {
			return nil, ErrCartNotFound
		} else if err == postgres.ErrUserNotFound {
			return nil, ErrUserNotFound
		} else if err == postgres.ErrProductNotFound {
			return nil, ErrProductNotFound
		} else if err == postgres.ErrDefaultPriceListNotFound {
			return nil, ErrDefaultPriceListNotFound
		} else if err == postgres.ErrCartProductExists {
			return nil, ErrCartProductExists
		} else if err == postgres.ErrProductHasNoPrices {
			return nil, ErrProductHasNoPrices
		}
		return nil, errors.Wrapf(err, "s.model.AddProductToCart(ctx, cartID=%q, %q, productID=%q, qty=%d) failed: ", cartID, "default", productID, qty)
	}
	sitem := CartProduct{
		Object:    "cart_product",
		ID:        item.UUID,
		CartID:    cartID,
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

// HasCartProducts returns true if any cart products have previously
// been added to the given cart.
func (s *Service) HasCartProducts(ctx context.Context, id string) (bool, error) {
	has, err := s.model.HasCartProducts(ctx, id)
	if err != nil {
		return false, errors.Wrap(err, "service: has cart items failed")
	}
	return has, nil
}

// GetCartProducts get all cart items by cart ID.
func (s *Service) GetCartProducts(ctx context.Context, cartID string) ([]*CartProduct, error) {
	userUUID := ctx.Value("ecom_uid").(string)

	cartProducts, err := s.model.GetCartProducts(ctx, cartID, userUUID)
	if err != nil {
		if err == postgres.ErrCartNotFound {
			return nil, ErrCartNotFound
		}
		return nil, err
	}

	results := make([]*CartProduct, 0, 32)
	for _, v := range cartProducts {
		i := CartProduct{
			Object:    "cart_product",
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

// UpdateCartProduct updates a cart item quantity.
func (s *Service) UpdateCartProduct(ctx context.Context, userID, cartProductID string, qty int) (*CartProduct, error) {
	item, err := s.model.UpdateCartProduct(ctx, userID, cartProductID, qty)
	if err != nil {
		if err == postgres.ErrCartNotFound {
			return nil, ErrCartNotFound
		} else if err == postgres.ErrProductNotFound {
			return nil, ErrProductNotFound
		} else if err == postgres.ErrCartProductNotFound {
			return nil, ErrCartProductNotFound
		}
		return nil, err
	}
	sitem := CartProduct{
		Object:    "cart_product",
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

// DeleteCartProduct deletes a single cart product.
func (s *Service) DeleteCartProduct(ctx context.Context, cartProductID string) error {
	if err := s.model.DeleteCartProduct(ctx, cartProductID); err != nil {
		if err == postgres.ErrCartProductNotFound {
			return ErrCartProductNotFound
		}
		return errors.Wrapf(err, "s.model.DeleteCartProduct(ctx, cartProductID=%q) failed", cartProductID)
	}
	return nil
}

// EmptyCartProducts empties the cart of all product (not including coupons).
func (s *Service) EmptyCartProducts(ctx context.Context, cartID string) error {
	err := s.model.EmptyCartProducts(ctx, cartID)
	if err != nil {
		if err == postgres.ErrCartNotFound {
			return ErrCartNotFound
		}
		return err
	}
	return nil
}
