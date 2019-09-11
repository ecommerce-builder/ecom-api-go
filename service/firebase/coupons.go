package firebase

import (
	"context"
	"time"

	"bitbucket.org/andyfusniakteam/ecom-api-go/model/postgres"
	"github.com/pkg/errors"
)

// ErrCouponExists coupon already exists
var ErrCouponExists = errors.New("service: coupon exists")

// ErrCouponNotFound coupon not found
var ErrCouponNotFound = errors.New("service: coupon not found")

// ErrCouponInUse error
var ErrCouponInUse = errors.New("service: coupon in use")

// Coupon a single coupon for use with the cart.
type Coupon struct {
	Object      string    `json:"object"`
	ID          string    `json:"id"`
	CouponCode  string    `json:"coupon_code"`
	PromoRuleID string    `json:"promo_rule_id"`
	Void        bool      `json:"void"`
	Resuable    bool      `json:"resuable"`
	SpendCount  int       `json:"spend_count"`
	Created     time.Time `json:"created"`
	Modified    time.Time `json:"modfied"`
}

// CreateCoupon mints a new coupon to be later applied to a cart.
func (s *Service) CreateCoupon(ctx context.Context, couponCode, promoRuleID string, reusable bool) (*Coupon, error) {
	c, err := s.model.CreateCoupon(ctx, couponCode, promoRuleID, reusable)
	if err != nil {
		if err == postgres.ErrCouponExists {
			return nil, ErrCouponExists
		} else if err == postgres.ErrPromoRuleNotFound {
			return nil, ErrPromoRuleNotFound
		}
		return nil, errors.Wrapf(err, "service: s.model.CreateCoupon(ctx, couponCode=%q, promoRuleID=%q, reusable=%t)", couponCode, promoRuleID, reusable)
	}
	coupon := Coupon{
		Object:      "coupon",
		ID:          c.UUID,
		CouponCode:  c.CouponCode,
		PromoRuleID: c.PromoRuleUUID,
		Void:        c.Void,
		Resuable:    c.Resuable,
		SpendCount:  c.SpendCount,
		Created:     c.Created,
		Modified:    c.Modified,
	}
	return &coupon, nil
}

// GetCoupon returns a single coupon. If the coupon is not found
// returns nil with an error of `ErrCouponNotFound`.
func (s *Service) GetCoupon(ctx context.Context, couponID string) (*Coupon, error) {
	c, err := s.model.GetCouponByUUID(ctx, couponID)
	if err != nil {
		if err == postgres.ErrCouponNotFound {
			return nil, ErrCouponNotFound
		}
		return nil, errors.Wrapf(err, "service: s.model.GetCouponByUUID(ctx, couponID=%q)", couponID)
	}
	coupon := Coupon{
		Object:      "coupon",
		ID:          c.UUID,
		CouponCode:  c.CouponCode,
		PromoRuleID: c.PromoRuleUUID,
		Void:        c.Void,
		Resuable:    c.Resuable,
		SpendCount:  c.SpendCount,
		Created:     c.Created,
		Modified:    c.Modified,
	}
	return &coupon, nil
}

// GetCoupons returns a list coupons.
func (s *Service) GetCoupons(ctx context.Context) ([]*Coupon, error) {
	prows, err := s.model.GetCoupons(ctx)
	if err != nil {
		return nil, errors.Wrapf(err, "s.model.GetCoupons(ctx) failed")
	}

	coupons := make([]*Coupon, 0, len(prows))
	for _, c := range prows {
		coupon := Coupon{
			Object:      "coupon",
			ID:          c.UUID,
			CouponCode:  c.CouponCode,
			PromoRuleID: c.PromoRuleUUID,
			Void:        c.Void,
			Resuable:    c.Resuable,
			SpendCount:  c.SpendCount,
			Created:     c.Created,
			Modified:    c.Modified,
		}
		coupons = append(coupons, &coupon)
	}
	return coupons, nil
}

// UpdateCoupon partially updates am existing coupon. Returns
// either the updated coupon or nil with an error of `ErrCouponNotFound`.
func (s *Service) UpdateCoupon(ctx context.Context, couponID string, void *bool) (*Coupon, error) {
	c, err := s.model.UpdateCouponByUUID(ctx, couponID, void)
	if err != nil {
		if err == postgres.ErrCouponNotFound {
			return nil, ErrCouponNotFound
		}
		return nil, errors.Wrapf(err, "s.model.UpdateCoupon(ctx, couponID=%q, void=%v)", couponID, void)
	}

	coupon := Coupon{
		Object:      "coupon",
		ID:          c.UUID,
		CouponCode:  c.CouponCode,
		PromoRuleID: c.PromoRuleUUID,
		Void:        c.Void,
		Resuable:    c.Resuable,
		SpendCount:  c.SpendCount,
		Created:     c.Created,
		Modified:    c.Modified,
	}
	return &coupon, nil
}

// DeleteCoupon deletes an existing coupon.
func (s *Service) DeleteCoupon(ctx context.Context, couponID string) error {
	err := s.model.DeleteCouponByUUID(ctx, couponID)
	if err != nil {
		if err == postgres.ErrCouponNotFound {
			return ErrCouponNotFound
		} else if err == postgres.ErrCouponInUse {
			return ErrCouponInUse
		}
		return errors.Wrapf(err, "s.model.DeleteCouponByUUID(ctx, couponUUID=%q) failed", couponID)
	}
	return nil
}
