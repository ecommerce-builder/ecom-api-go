package postgres

import (
	"context"
	"database/sql"
	"time"

	"github.com/pkg/errors"
)

// ErrOfferNotFound error
var ErrOfferNotFound = errors.New("postgres: offer not found")

// OfferJoinRow joins between the offer and promo_rule table
type OfferJoinRow struct {
	id            int
	UUID          string
	promoRuleID   int
	PromoRuleUUID string
	Created       time.Time
	Modified      time.Time
}

// AddOffer adds an offer row to the offer table.
func (m *PgModel) AddOffer(ctx context.Context, promoRuleUUID string) (*OfferJoinRow, error) {
	// 1. Check the promo rule exists
	q1 := "SELECT id FROM promo_rule WHERE uuid = $1"
	var productRuleID int
	err := m.db.QueryRowContext(ctx, q1, promoRuleUUID).Scan(&productRuleID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrPromoRuleNotFound
		}
		return nil, errors.Wrapf(err, "postgres: query row context failed for q1=%q", q1)
	}

	// 2. Insert the offer
	q2 := `
		INSERT INTO offer (promo_rule_id, created, modified)
		VALUES ($1, NOW(), NOW())
		RETURNING id, uuid, promo_rule_id, created, modified
	`
	o := OfferJoinRow{}
	row := m.db.QueryRowContext(ctx, q2, productRuleID)
	if err := row.Scan(&o.id, &o.UUID, &o.promoRuleID, &o.Created, &o.Modified); err != nil {
		return nil, errors.Wrapf(err, "scan failed q2=%q", q2)
	}
	o.PromoRuleUUID = promoRuleUUID

	return &o, nil
}

// GetOfferByUUID return a single offer
func (m *PgModel) GetOfferByUUID(ctx context.Context, offerUUID string) (*OfferJoinRow, error) {
	q1 := `
		SELECT
		  o.id, o.uuid, promo_rule_id, r.uuid as promo_rule_uuid,
		  o.created, o.modified
		FROM offer AS o
		INNER JOIN promo_rule AS r
		  ON r.id = o.promo_rule_id
		WHERE o.uuid = $1
	`
	o := OfferJoinRow{}
	row := m.db.QueryRowContext(ctx, q1, offerUUID)
	if err := row.Scan(&o.id, &o.UUID, &o.promoRuleID, &o.PromoRuleUUID, &o.Created, &o.Modified); err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrOfferNotFound
		}
		return nil, errors.Wrapf(err, "query row context scan q1=%q", q1)
	}
	return &o, nil
}

// GetOffers returns a list of offers
func (m *PgModel) GetOffers(ctx context.Context) ([]*OfferJoinRow, error) {
	q1 := `
		SELECT
		  o.id, o.uuid, promo_rule_id, r.uuid as promo_rule_uuid,
		  o.created, o.modified
		FROM offer AS o
		INNER JOIN promo_rule AS r
		  ON r.id = o.promo_rule_id
	`
	rows, err := m.db.QueryContext(ctx, q1)
	if err != nil {
		return nil, errors.Wrapf(err, "m.db.QueryContext(ctx) failed")
	}
	defer rows.Close()

	offers := make([]*OfferJoinRow, 0, 2)
	for rows.Next() {
		var o OfferJoinRow
		if err = rows.Scan(&o.id, &o.UUID, &o.promoRuleID, &o.PromoRuleUUID, &o.Created, &o.Modified); err != nil {
			return nil, errors.Wrap(err, "scan failed")
		}
		offers = append(offers, &o)
	}
	if err = rows.Err(); err != nil {
		return nil, errors.Wrap(err, "rows.Err()")
	}
	return offers, nil
}

// DeleteOfferByUUID deletes an offer.
func (m *PgModel) DeleteOfferByUUID(ctx context.Context, offerUUID string) error {
	// 1. Check the offer exists
	q1 := "SELECT id FROM offer WHERE uuid = $1"
	var offerID int
	err := m.db.QueryRowContext(ctx, q1, offerUUID).Scan(&offerID)
	if err != nil {
		if err == sql.ErrNoRows {
			return ErrOfferNotFound
		}
		return errors.Wrapf(err, "postgres: query row context failed for q1=%q", q1)
	}

	// 2. Delete the offer
	q2 := "DELETE FROM offer WHERE id = $1"
	_, err = m.db.ExecContext(ctx, q2, offerID)
	if err != nil {
		return errors.Wrapf(err, "postgres: exec context q2=%q", q2)
	}

	return nil
}
