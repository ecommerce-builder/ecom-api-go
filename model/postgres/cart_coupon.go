package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"time"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

// ErrCouponExpired error
var ErrCouponExpired = errors.New("postgres: coupon expired")

// ErrCouponNotAtStartDate error
var ErrCouponNotAtStartDate = errors.New("postgres: coupon not at start date")

// ErrCouponVoid error
var ErrCouponVoid = errors.New("postgres: coupon voided")

// ErrCouponUsed error
var ErrCouponUsed = errors.New("postgres: coupon used")

// ErrCartCouponNotFound error
var ErrCartCouponNotFound = errors.New("postgres: cart coupon not found")

// ErrCartCouponExists error
var ErrCartCouponExists = errors.New("postgres: cart coupon exists")

// CartCouponJoinRow hold a single row from the join between the
// cart_coupon, cart and coupon table.
type CartCouponJoinRow struct {
	id            int
	UUID          string
	cartID        int
	CartUUID      string
	couponID      int
	CouponUUID    string
	CouponCode    string
	PromoRuleUUID string
	Created       time.Time
	Modified      time.Time
}

var loc *time.Location

func init() {
	var err error
	loc, err = time.LoadLocation("Europe/London")
	if err != nil {
		fmt.Fprintf(os.Stderr, "time.LoadLocation(%q) failed: %+v", "Europe/London", err.Error())
		return
	}
}

// AddCartCoupon adds a new row to the cart_coupon table.
func (m *PgModel) AddCartCoupon(ctx context.Context, cartUUID, couponUUID string) (*CartCouponJoinRow, error) {
	contextLogger := log.WithContext(ctx)
	contextLogger.Debugf("postgres: AddCartCoupon(ctx context.Context, cartUUID=%q, couponUUID=%q string) started", cartUUID, couponUUID)

	// 1. Check the cart exists
	q1 := "SELECT id FROM cart WHERE uuid = $1"
	var cartID int
	err := m.db.QueryRowContext(ctx, q1, cartUUID).Scan(&cartID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrCartNotFound
		}
		return nil, errors.Wrapf(err, "postgres: query row context failed for q1=%q", q1)
	}

	// 2. Check the coupon exists
	q2 := `
		SELECT
		  c.id, c.coupon_code, void, reusable, spend_count, r.uuid as promo_rule_uuid,
		  r.start_at, r.end_at
		FROM coupon AS c
		INNER JOIN promo_rule AS r
		  ON r.id = c.promo_rule_id
		WHERE c.uuid = $1
	`
	var void bool
	var reusable bool
	var spendCount int
	var startAt *time.Time
	var endAt *time.Time

	var c CartCouponJoinRow
	c.cartID = cartID
	c.CartUUID = cartUUID
	c.CouponUUID = couponUUID
	err = m.db.QueryRowContext(ctx, q2, couponUUID).Scan(&c.couponID, &c.CouponCode, &void, &reusable, &spendCount, &c.PromoRuleUUID, &startAt, &endAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrCouponNotFound
		}
		return nil, errors.Wrapf(err, "postgres: query row context failed for q2=%q", q2)
	}

	// 3. Check the coupon has not already been applied to the cart.
	q3 := "SELECT EXISTS(SELECT 1 FROM cart_coupon WHERE cart_id = $1 AND coupon_id = $2) AS exists"
	var exists bool
	err = m.db.QueryRowContext(ctx, q3, cartID, c.couponID).Scan(&exists)
	if err != nil {
		return nil, errors.Wrapf(err, "postgres:  m.db.QueryRowContext(ctx, q3=%q, cartID=%d, c.couponID=%d).Scan(...) failed", q3, cartID, c.couponID)
	}

	if exists {
		return nil, ErrCartCouponExists
	}

	// Check if the coupon has been voided
	if void {
		contextLogger.Debugf("postgres: coupon is void (couponUUID=%q)", couponUUID)
		return nil, ErrCouponVoid
	}

	// Check if the coupon has been used. (Only applies to non-reusable coupons)
	if !reusable && spendCount > 0 {
		contextLogger.Debug("postgres: coupon is not reusable and spendCount > 0. The coupon has been already used.")
		return nil, ErrCouponUsed
	}

	// Check if the coupon has expired
	if startAt != nil && endAt != nil {
		now := time.Now()

		diffStart := now.Sub(*startAt)

		// if the difference is a negative value,
		// then the coupon hasn't yet started
		hoursToStart := diffStart.Hours()
		if hoursToStart < 0 {
			contextLogger.Infof("postgres: coupon is %.1f hours before start at date", hoursToStart)
			return nil, ErrCouponNotAtStartDate
		}

		// if the difference is a positive value,
		// then the coupon has expired.
		diffEnd := now.Sub(*endAt)
		hoursOverEnd := diffEnd.Hours()
		if hoursOverEnd > 0 {
			contextLogger.Infof("postgres: coupon is %.1f hours over the end at date", hoursOverEnd)
			return nil, ErrCouponExpired
		}

		format := "2006-01-02 15:04:05 GMT"
		contextLogger.Infof("postgres: coupon is between %s and %s.", startAt.In(loc).Format(format), endAt.In(loc).Format(format))
	} else {
		contextLogger.Debugf("postgres: startAt for promo rule %q is nil", c.PromoRuleUUID)
	}

	// 4. Insert the cart coupon.
	q4 := `
		INSERT INTO cart_coupon
		  (cart_id, coupon_id)
		VALUES
		  ($1, $2)
		RETURNING
		  id, uuid, cart_id, coupon_id, created, modified
	`
	row := m.db.QueryRowContext(ctx, q4, cartID, c.couponID)
	if err := row.Scan(&c.id, &c.UUID, &c.cartID, &c.couponID, &c.Created, &c.Modified); err != nil {
		return nil, errors.Wrapf(err, "postgres: query scan failed q4=%q", q4)
	}

	return &c, nil
}

// GetCartCoupon retrieves a single cart_coupon row by uuid. If the row
// does not exist returns nil, and a `ErrCartCouponNotFound` error.
func (m *PgModel) GetCartCoupon(ctx context.Context, cartCouponUUID string) (*CartCouponJoinRow, error) {
	q1 := `
		SELECT
		  p.id, p.uuid, p.cart_id, c.uuid as cart_uuid, p.coupon_id,
		  u.uuid as coupon_uuid, u.coupon_code as coupon_code,
		  r.uuid as promo_rule_uuid, p.created, p.modified
		FROM cart_coupon AS p
		INNER JOIN cart AS c
		  ON c.id = p.cart_id
		INNER JOIN coupon AS u
		  ON u.id = p.coupon_id
		INNER JOIN promo_rule AS r
		  ON r.id = u.promo_rule_id
		WHERE p.uuid = $1
	`
	var c CartCouponJoinRow
	err := m.db.QueryRowContext(ctx, q1, cartCouponUUID).Scan(&c.id, &c.UUID, &c.cartID, &c.CartUUID, &c.couponID,
		&c.CouponUUID, &c.CouponCode, &c.PromoRuleUUID, &c.Created, &c.Modified)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrCartCouponNotFound
		}
		return nil, errors.Wrapf(err, "postgres: query row context failed for q1=%q", q1)
	}
	return &c, nil
}

// GetCartCouponsByCartUUID returns all cart coupons for a given cart uuid.
func (m *PgModel) GetCartCouponsByCartUUID(ctx context.Context, cartUUID string) ([]*CartCouponJoinRow, error) {
	// 1. Check the cart exists
	q1 := "SELECT id FROM cart WHERE uuid = $1"
	var cartID int
	err := m.db.QueryRowContext(ctx, q1, cartUUID).Scan(&cartID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrCartNotFound
		}
		return nil, errors.Wrapf(err, "postgres: query row context failed for q1=%q", q1)
	}

	// 2. Get all the coupons applied to this cart.
	q2 := `
		SELECT
		  p.id, p.uuid, p.cart_id, c.uuid as cart_uuid, p.coupon_id,
		  u.uuid as coupon_uuid, u.coupon_code as coupon_code,
		  r.uuid as promo_rule_uuid, p.created, p.modified
		FROM cart_coupon AS p
		INNER JOIN cart AS c
		  ON c.id = p.cart_id
		INNER JOIN coupon AS u
		  ON u.id = p.coupon_id
		INNER JOIN promo_rule AS r
		  ON r.id = u.promo_rule_id
		WHERE p.cart_id = $1
	`
	rows, err := m.db.QueryContext(ctx, q2, cartID)
	if err != nil {
		return nil, errors.Wrapf(err, "postgres: m.db.QueryContext(ctx) failed")
	}
	defer rows.Close()

	cartCoupons := make([]*CartCouponJoinRow, 0)
	for rows.Next() {
		var c CartCouponJoinRow
		if err = rows.Scan(&c.id, &c.UUID, &c.cartID, &c.CartUUID, &c.couponID,
			&c.CouponUUID, &c.CouponCode, &c.PromoRuleUUID, &c.Created, &c.Modified); err != nil {
			if err == sql.ErrNoRows {
				return nil, ErrPriceListNotFound
			}
			return nil, errors.Wrap(err, "postgres: scan failed")
		}
		cartCoupons = append(cartCoupons, &c)
	}
	if err = rows.Err(); err != nil {
		return nil, errors.Wrap(err, "postgres: rows.Err()")
	}
	return cartCoupons, nil

}

// DeleteCartCoupon deletes a single cart_coupon row from the cart_coupon table.
func (m *PgModel) DeleteCartCoupon(ctx context.Context, cartCouponUUID string) error {
	// 1. Check the cart_coupon exists
	q1 := "SELECT id FROM cart_coupon WHERE uuid = $1"
	var cartCouponID int
	err := m.db.QueryRowContext(ctx, q1, cartCouponUUID).Scan(&cartCouponID)
	if err != nil {
		if err == sql.ErrNoRows {
			return ErrCartCouponNotFound
		}
		return errors.Wrapf(err, "postgres: query row context failed for q1=%q", q1)
	}

	// 2. Delete the cart_coupon
	q2 := "DELETE FROM cart_coupon WHERE id = $1"
	if _, err = m.db.ExecContext(ctx, q2, cartCouponID); err != nil {
		return errors.Wrapf(err, "model: delete category q2=%q", q2)
	}
	return nil
}
