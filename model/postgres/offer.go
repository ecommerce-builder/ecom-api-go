package postgres

import (
	"context"
	"database/sql"
	"math"
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
	if err == sql.ErrNoRows {
		return nil, ErrPromoRuleNotFound
	}
	if err != nil {
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

	tx, err := m.db.BeginTx(ctx, nil)
	if err != nil {
		return errors.Wrap(err, "postgres: db.BeginTx")
	}
	contextLogger.Debugf("postgres: db begin transaction")

	// 1. Get a list of all promo rules applied to offers.
	q1 := `
		SELECT
		  r.id, r.uuid, promo_rule_code, product_id, product_set_id,
		  category_id, shipping_tariff_id, name, start_at, end_at,
		  amount, total_threshold, type, target, r.created, r.modified
		FROM promo_rule AS r
		INNER JOIN offer AS o
		  ON o.promo_rule_id = r.id
	`
	rows, err := tx.QueryContext(ctx, q1)
	if err != nil {
		return errors.Wrapf(err, "postgres: tx.QueryContext(ctx) failed")
	}
	defer rows.Close()

	promos := make([]*PromoRuleRow, 0, 2)
	for rows.Next() {
		var p PromoRuleRow
		if err = rows.Scan(&p.id, &p.UUID, &p.PromoRuleCode, &p.productID, &p.productSetID, &p.categoryID, &p.shippingTariffID, &p.Name, &p.StartAt, &p.EndAt, &p.Amount, &p.TotalThreshold, &p.Type, &p.Target, &p.Created, &p.Modified); err != nil {
			return errors.Wrap(err, "postgres: scan failed")
		}
		promos = append(promos, &p)
	}
	if err := rows.Err(); err != nil {
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
				contextLogger.Infof("postgres: offer promo %q is %.1f mins over the end at datetime... skipping", promo.PromoRuleCode, minsOverEnd)
				continue
			}

			contextLogger.Infof("postgres: offer promo is between %s and %s.", promo.StartAt.In(loc).Format(format), promo.EndAt.In(loc).Format(format))
		} else {
			contextLogger.Debugf("postgres: startAt for offer promo rule code %q is nil so no time constraints for this offer promo", promo.PromoRuleCode)
		}

		contextLogger.Infof("postgres: offer promo %q has a target of %q", promo.PromoRuleCode, promo.Target)
		if promo.Target == "category" {
			categories := make([]int, 0)

			// 2. Read the category to determine if its a leaf or non-leaf
			// Non-leafs should get all leaf category descendants.
			q2 := "SELECT id, lft, rgt FROM category WHERE id = $1"
			var categoryID, lft, rgt int
			err = tx.QueryRowContext(ctx, q2, *promo.categoryID).Scan(&categoryID, &lft, &rgt)
			if err == sql.ErrNoRows {
				return ErrCategoryNotFound
			}
			if err != nil {
				return errors.Wrapf(err, "postgres: query row context failed for q2=%q", q2)
			}

			if lft == rgt-1 {
				// leaf category
				categories = append(categories, categoryID)
			} else {
				// non-leaf category
				q3 := "SELECT * FROM category WHERE lft > $1 AND rgt < $2 AND lft = rgt-1"
				rows, err := tx.QueryContext(ctx, q3)
				if err != nil {
					return errors.Wrapf(err, "postgres: tx.QueryContext(ctx) failed")
				}
				defer rows.Close()

				leafCategories := make([]*CategoryRow, 0)
				for rows.Next() {
					var c CategoryRow
					if err = rows.Scan(&c.id, &c.UUID, &c.Segment, &c.Path, &c.Name, &c.Lft, &c.Rgt, &c.Depth, &c.Created, &c.Modified); err != nil {
						return errors.Wrap(err, "postgres: scan failed")
					}
					leafCategories = append(leafCategories, &c)
					categories = append(categories, c.id)
				}
				if err := rows.Err(); err != nil {
					return errors.Wrap(err, "postgres: rows.Err()")
				}
			}

			// build a list of products for the categories
			q4 := `
				SELECT
				  id, uuid, product_id, category_id, pri, created, modified
				FROM product_category
				WHERE category_id IN ($1)
			`
			rows, err := tx.QueryContext(ctx, q4)
			if err != nil {
				return errors.Wrapf(err, "postgres: tx.QueryContext(ctx) failed")
			}
			defer rows.Close()

			productCategoryList := make([]*ProductCategoryRow, 0)
			for rows.Next() {
				var c ProductCategoryRow
				if err = rows.Scan(&c.id, &c.UUID, &c.productID, &c.categoryID, &c.Pri, &c.Created, &c.Modified); err != nil {
					return errors.Wrap(err, "postgres: scan failed")
				}
				productCategoryList = append(productCategoryList, &c)
			}
			if err := rows.Err(); err != nil {
				return errors.Wrap(err, "postgres: rows.Err()")
			}

			// for each product, calculate the discounted price and store it
			// for every price break of that product.
			// fmt.Println("products...")

			// 5. Prepare get product
			q5 := `
				SELECT
				  id, uuid, product_id, price_list_id, break, unit_price, created, modified
				FROM price
				WHERE product_id = $1
			`
			stmt3, err := tx.PrepareContext(ctx, q5)
			if err != nil {
				tx.Rollback()
				return errors.Wrapf(err, "postgres: tx prepare for q5=%q", q5)
			}
			defer stmt3.Close()

			// 6. Preprare price update
			q6 := `
				UPDATE price
				SET offer_price = $1, modified = NOW()
				WHERE id = $2
			`
			stmt6, err := tx.PrepareContext(ctx, q6)
			if err != nil {
				tx.Rollback()
				return errors.Wrapf(err, "postgres: tx prepare for q6=%q", q6)
			}
			defer stmt6.Close()

			for _, productID := range productCategoryList {
				// fmt.Printf("%+v\n", productID)

				rows, err := stmt3.QueryContext(ctx, productID)
				if err != nil {
					tx.Rollback()
					return errors.Wrapf(err, "postgres: stmt3.QueryContext(ctx, ...) failed q6=%q", q6)
				}
				defer rows.Close()

				prices := make([]*PriceRow, 0, 1)
				for rows.Next() {
					var p PriceRow
					if err = rows.Scan(&p.id, &p.UUID, &p.productID, &p.priceListID, &p.Break, &p.UnitPrice, &p.Created, &p.Modified); err != nil {
						return errors.Wrap(err, "postgres: scan failed")
					}
					prices = append(prices, &p)
				}
				if err := rows.Err(); err != nil {
					return errors.Wrap(err, "postgres: rows.Err()")
				}

				for _, price := range prices {
					var offerPrice int
					if promo.Type == "fixed" {
						contextLogger.Debugf("postgres: promo type is fixed with amount %d", promo.Amount)
						offerPrice = price.UnitPrice - promo.Amount
					} else if promo.Type == "percentage" {
						contextLogger.Debugf("postgres: promo type is percentage with amount %d", promo.Amount)
						// percentage is stored as a integer between 0 to 10,000
						// 0 = 0.00% and 9,999 = 99.99% discount.
						fraction := float64(promo.Amount) / 10000.0
						offerPrice = int(math.Round(float64(price.UnitPrice) * fraction))
					}

					contextLogger.Infof("postgres: offerPrice=%d (price uuid=%q productID=%d priceListID=%d Break=%d UnitPrice=%d)", offerPrice, price.UUID, price.productID, price.priceListID, price.Break, price.UnitPrice)
					_, err := stmt6.ExecContext(ctx, offerPrice, price.id)
					if err != nil {
						return errors.Wrapf(err, "postgres: stmt4.ExecContext(ctx, offerPrice=%d, price.id=%d) failed", offerPrice, price.id)
					}
				}
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return errors.Wrap(err, "postgres: tx.Commit")
	}
	contextLogger.Debugf("postgres: db commit succeeded")

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
	err := row.Scan(&o.id, &o.UUID, &o.promoRuleID, &o.PromoRuleUUID, &o.Created, &o.Modified)
	if err == sql.ErrNoRows {
		return nil, ErrOfferNotFound
	}
	if err != nil {
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
	if err := rows.Err(); err != nil {
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
	if err == sql.ErrNoRows {
		return ErrOfferNotFound
	}
	if err != nil {
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
