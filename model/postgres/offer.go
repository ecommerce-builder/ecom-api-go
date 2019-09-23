package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

// ErrOfferNotFound error
var ErrOfferNotFound = errors.New("postgres: offer not found")

// ErrOfferExists error
var ErrOfferExists = errors.New("postgres: offer exists")

// OfferRow holds data from a single row in the offer table.
type OfferRow struct {
	id          int
	UUID        string
	promoRuleID int
	Created     time.Time
	Modified    time.Time
}

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

	// 2. Check if the offer is already activated.
	q2 := "SELECT EXISTS(SELECT 1 FROM offer WHERE promo_rule_id = $1) AS exists"
	var exists bool
	if err := m.db.QueryRowContext(ctx, q2, productRuleID).Scan(&exists); err != nil {
		return nil, errors.Wrapf(err, "postgres: tx.QueryRowContext(ctx, q2=%q, productRuleID=%d) failed", q2, productRuleID)
	}
	if exists {
		return nil, ErrOfferExists
	}

	// 3. Insert the offer
	q3 := `
		INSERT INTO offer (promo_rule_id, created, modified)
		VALUES ($1, NOW(), NOW())
		RETURNING id, uuid, promo_rule_id, created, modified
	`
	o := OfferJoinRow{}
	row := m.db.QueryRowContext(ctx, q3, productRuleID)
	if err := row.Scan(&o.id, &o.UUID, &o.promoRuleID, &o.Created, &o.Modified); err != nil {
		return nil, errors.Wrapf(err, "scan failed q3=%q", q3)
	}
	o.PromoRuleUUID = promoRuleUUID

	return &o, nil
}

// CalcOfferPrices resolves the offer for each product.
func (m *PgModel) CalcOfferPrices(ctx context.Context) error {
	contextLogger := log.WithContext(ctx)
	contextLogger.Debugf("postgres: CalcOfferPrices(ctx context.Context) started")

	// 1. Get a list of all promo rules applied to offers.
	q1 := `
		SELECT
		  r.id, r.uuid, promo_rule_code, product_id, product_set_id,
		  category_id, shipping_tarrif_id, name, start_at, end_at,
		  amount, total_threshold, type, target, r.created, r.modified
		FROM promo_rule AS r
		INNER JOIN offer AS o
		  ON o.promo_rule_id = r.id
	`
	rows, err := m.db.QueryContext(ctx, q1)
	if err != nil {
		return errors.Wrapf(err, "postgres: m.db.QueryContext(ctx) failed")
	}
	defer rows.Close()

	promos := make([]*PromoRuleRow, 0, 2)
	for rows.Next() {
		var p PromoRuleRow
		if err = rows.Scan(&p.id, &p.UUID, &p.PromoRuleCode, &p.productID, &p.productSetID, &p.categoryID, &p.shippingTarrifID, &p.Name, &p.StartAt, &p.EndAt, &p.Amount, &p.TotalThreshold, &p.Type, &p.Target, &p.Created, &p.Modified); err != nil {
			return errors.Wrap(err, "postgres: scan failed")
		}
		promos = append(promos, &p)
	}
	if err = rows.Err(); err != nil {
		return errors.Wrap(err, "postgres: rows.Err()")
	}

	// 2. For each promo rule, get a list of products that this promo applies to.
	now := time.Now()
	format := "2006-01-02 15:04:05 GMT"
	for _, promo := range promos {
		if promo.StartAt != nil && promo.EndAt != nil {
			diffStart := now.Sub(*promo.StartAt)
			// if the difference is a negative value,
			// then the offer hasn't yet started.
			minsToStart := diffStart.Minutes()
			if minsToStart < 0 {
				contextLogger.Infof("postgres: offer promo %q is %.1f mins before start_at datetime... skipping", promo.PromoRuleCode, minsToStart)
				continue
			}

			// if the difference is a positive value,
			// then the offer has expired.
			diffEnd := now.Sub(*promo.EndAt)
			minsOverEnd := diffEnd.Minutes()
			if minsOverEnd > 0 {
				contextLogger.Infof("postgres: offer promo %q is %.1f mins over the end at datetime... skipping", minsOverEnd)
				continue
			}

			contextLogger.Infof("postgres: offer promo is between %s and %s.", promo.StartAt.In(loc).Format(format), promo.EndAt.In(loc).Format(format))
		} else {
			contextLogger.Debugf("postgres: startAt for offer promo rule code %q is nil so no time constraints for this offer promo", promo.PromoRuleCode)
		}

		fmt.Printf("%#v\n", promo)

		contextLogger.Infof("postgres: offer promo %q has a target of %q", promo.PromoRuleCode, promo.Target)
		if promo.Target == "category" {
			// 2. Read the category to determine if its a leaf or non-leaf
			// Non-leafs should get all leaf category descendants.
			q2 := "SELECT lft, rgt FROM category WHERE id = $1"
			var lft, rgt int
			err = m.db.QueryRowContext(ctx, q2, *promo.categoryID).Scan(&lft, &rgt)
			if err != nil {
				if err == sql.ErrNoRows {
					return ErrCategoryNotFound
				}
				return errors.Wrapf(err, "postgres: query row context failed for q2=%q", q2)
			}

			// if this is a non-leaf category get a list of leaf categories
			if lft != rgt-1 {
				q3 := "SELECT * FROM category WHERE lft > 34 AND rgt < 119 AND lft = rgt-1"
				rows, err := m.db.QueryContext(ctx, q3)
				if err != nil {
					return errors.Wrapf(err, "postgres: m.db.QueryContext(ctx) failed")
				}
				defer rows.Close()

				leafCategories := make([]*CategoryRow, 0)
				for rows.Next() {
					var c CategoryRow
					if err = rows.Scan(&c.id, &c.UUID, &c.Segment, &c.Path, &c.Name, &c.Lft, &c.Rgt, &c.Depth, &c.Created, &c.Modified); err != nil {
						return errors.Wrap(err, "postgres: scan failed")
					}
					leafCategories = append(leafCategories, &c)
				}
				if err = rows.Err(); err != nil {
					return errors.Wrap(err, "postgres: rows.Err()")
				}
			}
		}
	}

	return nil
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
