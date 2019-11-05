package postgres

import (
	"context"
	"database/sql"
	"time"

	"github.com/pkg/errors"
)

// ErrCouponExists coupon already exists
var ErrCouponExists = errors.New("postgres: coupon exists")

// ErrCouponNotFound coupon not found
var ErrCouponNotFound = errors.New("postgres: coupon not found")

// ErrCouponInUse error
var ErrCouponInUse = errors.New("postgres: coupon in use")

// CouponJoinRow holds the join between the coupon and promo_rule table.
type CouponJoinRow struct {
	id            int
	UUID          string
	CouponCode    string
	promoRuleID   int
	PromoRuleUUID string
	PromoRuleCode string
	Void          bool
	Resuable      bool
	SpendCount    int
	Created       time.Time
	Modified      time.Time
}

// CreateCoupon adds a new coupon row to the coupon table.
func (m *PgModel) CreateCoupon(ctx context.Context, couponCode, promoRuleUUID string, reusable bool) (*CouponJoinRow, error) {
	// 1. Check if the coupon already exists
	q1 := `SELECT EXISTS(SELECT 1 FROM coupon WHERE coupon_code = $1) AS exists`
	var exists bool
	err := m.db.QueryRowContext(ctx, q1, couponCode).Scan(&exists)
	if err != nil {
		return nil, errors.Wrapf(err,
			"postgres: tx.QueryRowContext(ctx, q1=%q, couponCode=%q)",
			q1, couponCode)
	}
	if exists {
		return nil, ErrCouponExists
	}

	// 2. Check if the promo rule exists
	q2 := "SELECT id, promo_rule_code FROM promo_rule WHERE uuid = $1"
	var promoRuleID int
	var promoRuleCode string
	err = m.db.QueryRowContext(ctx, q2, promoRuleUUID).
		Scan(&promoRuleID, &promoRuleCode)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrPromoRuleNotFound
		}
		return nil, errors.Wrapf(err,
			"postgres: query row context failed for q2=%q", q2)
	}

	// 3. Insert the new coupon
	q3 := `
		INSERT INTO coupon
		  (coupon_code, promo_rule_id, reusable, spend_count, created, modified)
		VALUES ($1, $2, $3, 0, NOW(), NOW())
		RETURNING
		  id, uuid, coupon_code, promo_rule_id, void, reusable, spend_count, created, modified
	`
	c := CouponJoinRow{}
	row := m.db.QueryRowContext(ctx, q3, couponCode, promoRuleID, reusable)
	if err := row.Scan(&c.id, &c.UUID,
		&c.CouponCode, &c.promoRuleID,
		&c.Void, &c.Resuable, &c.SpendCount, &c.Created, &c.Modified); err != nil {
		return nil, errors.Wrapf(err, "query row context q3=%q", q3)
	}
	c.PromoRuleUUID = promoRuleUUID
	c.PromoRuleCode = promoRuleCode
	return &c, nil
}

// GetCouponByUUID returns a single coupon by uuid.
func (m *PgModel) GetCouponByUUID(ctx context.Context, couponUUID string) (*CouponJoinRow, error) {
	q1 := `
		SELECT
		  c.id, c.uuid,
		  coupon_code, promo_rule_id, r.uuid as promo_rule_uuid, r.promo_rule_code,
		  void, reusable, spend_count, c.created, c.modified
		FROM coupon AS c
		INNER JOIN promo_rule AS r
		  ON r.id = c.promo_rule_id
		WHERE c.uuid = $1
	`
	var c CouponJoinRow
	err := m.db.QueryRowContext(ctx, q1, couponUUID).Scan(&c.id, &c.UUID,
		&c.CouponCode, &c.promoRuleID, &c.PromoRuleUUID, &c.PromoRuleCode,
		&c.Void, &c.Resuable, &c.SpendCount, &c.Created, &c.Modified)
	if err == sql.ErrNoRows {
		return nil, ErrCouponNotFound
	}
	if err != nil {
		return nil, errors.Wrapf(err, "postgres: query row context failed for q1=%q", q1)
	}
	return &c, nil
}

// GetCoupons returns a list of CouponJoinRows.
func (m *PgModel) GetCoupons(ctx context.Context) ([]*CouponJoinRow, error) {
	q1 := `
		SELECT
		  c.id, c.uuid,
		  coupon_code, promo_rule_id, r.uuid as promo_rule_uuid, r.promo_rule_code,
		  void, reusable, spend_count, c.created, c.modified
		FROM coupon AS c
		INNER JOIN promo_rule AS r
		  ON r.id = c.promo_rule_id
	`
	rows, err := m.db.QueryContext(ctx, q1)
	if err != nil {
		return nil, errors.Wrapf(err, "postgres: m.db.QueryContext(ctx) failed")
	}
	defer rows.Close()

	coupons := make([]*CouponJoinRow, 0, 4)
	for rows.Next() {
		var c CouponJoinRow
		err = rows.Scan(&c.id, &c.UUID,
			&c.CouponCode, &c.promoRuleID, &c.PromoRuleUUID, &c.PromoRuleCode,
			&c.Void, &c.Resuable, &c.SpendCount, &c.Created, &c.Modified)
		if err != nil {
			return nil, errors.Wrap(err, "postgres: scan failed")
		}
		coupons = append(coupons, &c)
	}
	if err := rows.Err(); err != nil {
		return nil, errors.Wrap(err, "postgres: rows.Err()")
	}
	return coupons, nil
}

// UpdateCouponByUUID updates a coupon by uuid.
func (m *PgModel) UpdateCouponByUUID(ctx context.Context, couponUUID string, void *bool) (*CouponJoinRow, error) {
	// 1. Check if the coupon exists
	q1 := `
		SELECT c.id, r.uuid as promo_rule_uuid, r.promo_rule_code
		FROM coupon AS c
		INNER JOIN promo_rule AS r
		 ON r.id = c.promo_rule_id
		WHERE c.uuid = $1
	`
	var couponID int
	var promoRuleUUID string
	var promoRuleCode string
	err := m.db.QueryRowContext(ctx, q1, couponUUID).Scan(
		&couponID, &promoRuleUUID, &promoRuleCode)
	if err == sql.ErrNoRows {
		return nil, ErrCouponNotFound
	}
	if err != nil {
		return nil, errors.Wrapf(err,
			"postgres: tx.QueryRowContext(ctx, q1=%q, couponUUID=%q)",
			q1, couponUUID)
	}

	// 2. Update the coupon
	q2 := `
		UPDATE coupon
		SET
		  void = $1, modified = NOW()
		WHERE id = $2
		RETURNING
		  id, uuid, coupon_code, promo_rule_id, void, reusable, spend_count, created, modified
	`
	c := CouponJoinRow{}
	row := m.db.QueryRowContext(ctx, q2, *void, couponID)
	err = row.Scan(&c.id, &c.UUID, &c.CouponCode, &c.promoRuleID,
		&c.Void, &c.Resuable, &c.SpendCount, &c.Created, &c.Modified)
	if err != nil {
		return nil, errors.Wrapf(err, "postgres: scan q2=%q", q2)
	}
	c.PromoRuleUUID = promoRuleUUID
	c.PromoRuleCode = promoRuleCode
	return &c, nil
}

// DeleteCouponByUUID deletes a coupon row by uuid.
func (m *PgModel) DeleteCouponByUUID(ctx context.Context, couponUUID string) error {
	// 1. Check if the coupon exists
	q1 := "SELECT id FROM coupon WHERE uuid = $1"
	var couponID int
	err := m.db.QueryRowContext(ctx, q1, couponUUID).Scan(&couponID)
	if err == sql.ErrNoRows {
		return ErrCouponNotFound
	}
	if err != nil {
		return errors.Wrapf(err, "postgres: tx.QueryRowContext(ctx, q1=%q, couponUUID=%q)", q1, couponUUID)
	}

	// 2. Check if the coupon is in use.
	q2 := "SELECT COUNT(*) AS count FROM cart_coupon WHERE coupon_id = $1"
	var count int
	err = m.db.QueryRowContext(ctx, q2, couponID).Scan(&count)
	if err != nil {
		return errors.Wrapf(err, "postgres: tx.QueryRowContext(ctx, q2=%q, couponID=%d)", q2, couponID)
	}
	if count > 0 {
		return ErrCouponInUse
	}

	// 3. Delete the coupon
	q3 := "DELETE FROM coupon WHERE id = $1"
	_, err = m.db.ExecContext(ctx, q3, couponID)
	if err != nil {
		return errors.Wrapf(err, "model: exec context q3=%q", q3)
	}
	return nil
}
