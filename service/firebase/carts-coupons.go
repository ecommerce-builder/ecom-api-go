package firebase

import (
	"context"
	"time"

	"bitbucket.org/andyfusniakteam/ecom-api-go/model/postgres"
	"github.com/pkg/errors"
)

// ErrCouponExpired error
var ErrCouponExpired = errors.New("service: coupon expired")

// ErrCouponNotAtStartDate error
var ErrCouponNotAtStartDate = errors.New("service: coupon not at start date")

// ErrCouponVoid error
var ErrCouponVoid = errors.New("service: coupon voided")

// ErrCouponUsed error
var ErrCouponUsed = errors.New("service: coupon used")

// ErrCartCouponNotFound error
var ErrCartCouponNotFound = errors.New("service: cart coupon not found")

// ErrCartCouponExists error
var ErrCartCouponExists = errors.New("service: cart coupon exists")

// CartCoupon represents a coupon applied with a cart
type CartCoupon struct {
	Object      string    `json:"object"`
	ID          string    `json:"id"`
	CartID      string    `json:"cart_id"`
	CouponID    string    `json:"coupon_id"`
	CouponCode  string    `json:"coupon_code"`
	PromoRuleID string    `json:"promo_rule_id"`
	Created     time.Time `json:"created"`
	Modified    time.Time `json:"modified"`
}

// ApplyCouponToCart applies a coupon to a cart provided that coupon
// exists and has not expired, been voided or previously used (for a
// non-reusable coupon)
func (s *Service) ApplyCouponToCart(ctx context.Context, cartID, couponID string) (*CartCoupon, error) {
	prow, err := s.model.AddCartCoupon(ctx, cartID, couponID)
	if err != nil {
		if err == postgres.ErrCartNotFound {
			return nil, ErrCartNotFound
		} else if err == postgres.ErrCouponNotFound {
			return nil, ErrCouponNotFound
		} else if err == postgres.ErrCartCouponExists {
			return nil, ErrCartCouponExists
		} else if err == postgres.ErrCouponNotAtStartDate {
			return nil, ErrCouponNotAtStartDate
		} else if err == postgres.ErrCouponExpired {
			return nil, ErrCouponExpired
		} else if err == postgres.ErrCouponVoid {
			return nil, ErrCouponVoid
		} else if err == postgres.ErrCouponUsed {
			return nil, ErrCouponUsed
		}

		return nil, errors.Wrapf(err, "s.model.AddCartCoupon(ctx, cartID=%q, couponID=%q) failed", cartID, couponID)
	}

	cartCoupon := CartCoupon{
		Object:      "cart_coupon",
		ID:          prow.UUID,
		CartID:      prow.CartUUID,
		CouponID:    prow.CouponUUID,
		CouponCode:  prow.CouponCode,
		PromoRuleID: prow.PromoRuleUUID,
		Created:     prow.Created,
		Modified:    prow.Modified,
	}
	return &cartCoupon, nil
}

// GetCartCoupon retrieves a cart coupon relation by id.
func (s *Service) GetCartCoupon(ctx context.Context, cartCouponID string) (*CartCoupon, error) {
	row, err := s.model.GetCartCoupon(ctx, cartCouponID)
	if err != nil {
		if err == postgres.ErrCartCouponNotFound {
			return nil, ErrCartCouponNotFound
		}
		return nil, errors.Wrapf(err, "service: s.model.GetCartCoupon(ctx, cartCouponID=%q) failed", cartCouponID)
	}
	cartCoupon := CartCoupon{
		Object:      "cart_coupon",
		ID:          row.UUID,
		CartID:      row.CartUUID,
		CouponID:    row.CouponUUID,
		CouponCode:  row.CouponCode,
		PromoRuleID: row.PromoRuleUUID,
		Created:     row.Created,
		Modified:    row.Modified,
	}
	return &cartCoupon, nil
}

// GetCartCoupons returns a slice of cart coupons.
func (s *Service) GetCartCoupons(ctx context.Context, cartID string) ([]*CartCoupon, error) {
	prows, err := s.model.GetCartCouponsByCartUUID(ctx, cartID)
	if err != nil {
		if err == postgres.ErrCartNotFound {
			return nil, ErrCartNotFound
		}
		return nil, errors.Wrapf(err, "")
	}

	cartCoupons := make([]*CartCoupon, 0, len(prows))
	for _, row := range prows {
		c := CartCoupon{
			Object:      "cart_coupon",
			ID:          row.UUID,
			CartID:      row.CartUUID,
			CouponID:    row.CouponUUID,
			CouponCode:  row.CouponCode,
			PromoRuleID: row.PromoRuleUUID,
			Created:     row.Created,
			Modified:    row.Modified,
		}
		cartCoupons = append(cartCoupons, &c)
	}
	return cartCoupons, nil
}

// UnapplyCartCoupon unapplies a coupon from a cart.
func (s *Service) UnapplyCartCoupon(ctx context.Context, cartCouponID string) error {
	err := s.model.DeleteCartCoupon(ctx, cartCouponID)
	if err != nil {
		if err == postgres.ErrCartCouponNotFound {
			return ErrCartCouponNotFound
		}
		return errors.Wrapf(err, "service: s.model.DeleteCartCoupon(ctx, cartCouponID=%q) failed", cartCouponID)
	}
	return nil
}
