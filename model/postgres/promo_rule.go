package postgres

import (
	"context"
	"database/sql"
	"time"

	"github.com/pkg/errors"
)

// ErrPromoRuleNotFound error
var ErrPromoRuleNotFound = errors.New("model: promo rule not found")

// PromoRuleRow maps to a single row in the promo_rule table.
type PromoRuleRow struct {
	id             int
	UUID           string
	Name           string
	StartsAt       *time.Time
	EndsAt         *time.Time
	Amount         int
	TotalThreshold *int
	Type           string
	Target         string
	Created        time.Time
	Modified       time.Time
}

// PromoRuleCreate holds the required fields to create a new promo rule.
type PromoRuleCreate struct {
	Name           string
	StartsAt       *time.Time
	EndsAt         *time.Time
	Amount         int
	TotalThreshold *int
	Type           string
	Target         string
}

// CreatePromoRule creates a new promo rule row in the promo_rule table.
func (m *PgModel) CreatePromoRule(ctx context.Context, pr *PromoRuleCreate) (*PromoRuleRow, error) {
	q1 := `
		INSERT INTO promo_rule
		  (name, starts_at, ends_at, amount, total_threshold, type, target, created, modified)
		VALUES ($1, $2, $3, $4, $5, $6, $7, NOW(), NOW())
		RETURNING
		  id, uuid, name, starts_at, ends_at, amount, total_threshold, type, target, created, modified
	`
	r := PromoRuleRow{}
	row := m.db.QueryRowContext(ctx, q1, pr.Name, pr.StartsAt, pr.EndsAt, pr.Amount, pr.TotalThreshold, pr.Type, pr.Target)
	if err := row.Scan(&r.id, &r.UUID, &r.Name, &r.StartsAt, &r.EndsAt, &r.Amount, &r.TotalThreshold, &r.Type, &r.Target, &r.Created, &r.Modified); err != nil {
		return nil, errors.Wrapf(err, "query row context q1=%q", q1)
	}
	return &r, nil
}

// GetPromoRule returns a single promo rule row.
func (m *PgModel) GetPromoRule(ctx context.Context, promoRuleUUID string) (*PromoRuleRow, error) {
	q1 := `
		SELECT
		id, uuid, name, starts_at, ends_at, amount, total_threshold, type, target, created, modified
		FROM promo_rule WHERE uuid = $1
	`
	p := PromoRuleRow{}
	row := m.db.QueryRowContext(ctx, q1, promoRuleUUID)
	if err := row.Scan(&p.id, &p.UUID, &p.Name, &p.StartsAt, &p.EndsAt, &p.Amount, &p.TotalThreshold, &p.Type, &p.Target, &p.Created, &p.Modified); err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrPromoRuleNotFound
		}
		return nil, errors.Wrapf(err, "model: query row context scan query=%q promoRuleUUID=%q failed", q1, promoRuleUUID)
	}
	return &p, nil
}

// GetPromoRules returns a list of promo rules.
func (m *PgModel) GetPromoRules(ctx context.Context) ([]*PromoRuleRow, error) {
	q1 := `
		SELECT
		  id, uuid, name, starts_at, ends_at, amount, total_threshold, type, target, created, modified
		FROM promo_rule
	`
	rows, err := m.db.QueryContext(ctx, q1)
	if err != nil {
		return nil, errors.Wrapf(err, "model: m.db.QueryContext(ctx) failed")
	}
	defer rows.Close()

	rules := make([]*PromoRuleRow, 0, 4)
	for rows.Next() {
		var p PromoRuleRow
		if err = rows.Scan(&p.id, &p.UUID, &p.Name, &p.StartsAt, &p.EndsAt, &p.Amount, &p.TotalThreshold, &p.Type, &p.Target, &p.Created, &p.Modified); err != nil {
			if err == sql.ErrNoRows {
				return nil, ErrPromoRuleNotFound
			}
			return nil, errors.Wrap(err, "model: scan failed")
		}
		rules = append(rules, &p)
	}
	if err = rows.Err(); err != nil {
		return nil, errors.Wrap(err, "rows.Err()")
	}
	return rules, nil
}

// DeletePromoRule deletes a promo rule row from the promo_rule table.
func (m *PgModel) DeletePromoRule(ctx context.Context, promoRuleUUID string) error {
	tx, err := m.db.BeginTx(ctx, nil)
	if err != nil {
		return errors.Wrap(err, "model: db.BeginTx")
	}

	q1 := "SELECT id FROM promo_rule WHERE uuid = $1"
	var promoRuleID int
	err = tx.QueryRowContext(ctx, q1, promoRuleUUID).Scan(&promoRuleID)
	if err != nil {
		if err == sql.ErrNoRows {
			tx.Rollback()
			return ErrPromoRuleNotFound
		}
		tx.Rollback()
		return errors.Wrapf(err, "model: query row context failed for q1=%q", q1)
	}

	q2 := "DELETE FROM promo_rule WHERE id = $1"
	_, err = tx.ExecContext(ctx, q2, promoRuleID)
	if err != nil {
		tx.Rollback()
		return errors.Wrapf(err, "model: exec context q2=%q", q2)
	}

	if err = tx.Commit(); err != nil {
		return errors.Wrap(err, "model: tx.Commit")
	}
	return nil
}
