package postgres

import (
	"context"
	"database/sql"
	"time"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

// ErrPromoRuleNotFound error
var ErrPromoRuleNotFound = errors.New("postgres: promo rule not found")

// ErrPromoRuleExists error
var ErrPromoRuleExists = errors.New("postgres: promo rule exists")

// PromoRuleCreateProduct struct
type PromoRuleCreateProduct struct {
	ProductUUID string
}

// PromoRuleRow holds a single row in the promo_rule table.
type PromoRuleRow struct {
	id               int
	UUID             string
	PromoRuleCode    string
	productID        *int
	productSetID     *int
	categoryID       *int
	shippingTariffID *int
	Name             string
	StartAt          *time.Time
	EndAt            *time.Time
	Amount           int
	TotalThreshold   *int
	Type             string
	Target           string
	Created          time.Time
	Modified         time.Time
}

// PromoRuleJoinProductRow maps to a single row in the promo_rule table.
type PromoRuleJoinProductRow struct {
	id                 int
	UUID               string
	PromoRuleCode      string
	productID          *int
	ProductUUID        *string
	ProductPath        *string
	ProductSKU         *string
	productSetID       *int
	ProductSetUUID     *string
	categoryID         *int
	CategoryUUID       *string
	CategoryPath       *string
	shippingTariffID   *int
	ShippingTariffUUID *string
	ShippingTariffCode *string
	Name               string
	StartAt            *time.Time
	EndAt              *time.Time
	Amount             int
	TotalThreshold     *int
	Type               string
	Target             string
	Created            time.Time
	Modified           time.Time
}

// PromoRuleCreate holds the required fields to create a new promo rule.
type PromoRuleCreate struct {
	Name           string
	StartAt        *time.Time
	EndAt          *time.Time
	Amount         int
	TotalThreshold *int
	Type           string
	Target         string
}

// CreatePromoRuleTargetProduct creates a new promo rule row in the promo_rule table.
func (m *PgModel) CreatePromoRuleTargetProduct(ctx context.Context, productUUID, promoRuleCode, name string, startAt, endAt *time.Time, amount int, typ, target string) (*PromoRuleJoinProductRow, error) {
	contextLogger := log.WithContext(ctx)
	contextLogger.Debugf("postgres: CreatePromoRuleTargetProduct(ctx, productUUID=%q, promoRuleCode=%q, name=%q, startAt=%v, endAt=%v, amount=%d, typ=%q, target=%q) started", productUUID, promoRuleCode, name, startAt, endAt, amount, typ, target)

	tx, err := m.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, errors.Wrap(err, "postgres: db.BeginTx")
	}
	contextLogger.Debugf("postgres: db begin transaction")

	// 1. Check the product exists.
	q1 := "SELECT id, path, sku FROM product WHERE uuid = $1"
	var productID int
	var productPath string
	var productSKU string
	err = tx.QueryRowContext(ctx, q1, productUUID).Scan(&productID, &productPath, &productSKU)
	if err != nil {
		if err == sql.ErrNoRows {
			tx.Rollback()
			return nil, ErrProductNotFound
		}
		tx.Rollback()
		return nil, errors.Wrapf(err, "postgres: query row context failed for q1=%q", q1)
	}

	// 2. Check if the promo rule code exists.
	q2 := "SELECT EXISTS(SELECT 1 FROM promo_rule WHERE promo_rule_code = $1) AS exists"
	var exists bool
	if err := m.db.QueryRowContext(ctx, q2, promoRuleCode).Scan(&exists); err != nil {
		return nil, errors.Wrapf(err, "postgres: tx.QueryRowContext(ctx, q2=%q, promoRuleCode=%q) failed", q2, promoRuleCode)
	}
	if exists {
		contextLogger.Debugf("postgres: promo rule code %q already exists", promoRuleCode)
		return nil, ErrPromoRuleExists
	}
	contextLogger.Debugf("postgres: promo rule code %q is available", promoRuleCode)

	// 3. Create the promo rule with the product as its target.
	q3 := `
		INSERT INTO promo_rule
		  (product_id, promo_rule_code, name, start_at, end_at, amount, type, target, created, modified)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, NOW(), NOW())
		RETURNING
		  id, uuid, product_id, product_set_id, category_id, shipping_tariff_id,
		  name, start_at, end_at, amount, total_threshold, type, target, created, modified
	`
	r := PromoRuleJoinProductRow{}
	row := m.db.QueryRowContext(ctx, q3, productID, promoRuleCode, name, startAt, endAt, amount, typ, target)
	if err := row.Scan(&r.id, &r.UUID, &r.productID, &r.productSetID, &r.categoryID, &r.shippingTariffID, &r.Name, &r.StartAt, &r.EndAt, &r.Amount, &r.TotalThreshold, &r.Type, &r.Target, &r.Created, &r.Modified); err != nil {
		return nil, errors.Wrapf(err, "postgres: query row context q3=%q", q3)
	}
	r.ProductUUID = &productUUID
	r.ProductPath = &productPath
	r.ProductSKU = &productSKU

	if err = tx.Commit(); err != nil {
		return nil, errors.Wrap(err, "postgres: tx.Commit")
	}
	contextLogger.Debugf("postgres: db commit succeeded")
	return &r, nil
}

// CreatePromoRuleTargetProductSet create a promo rule associated to multiple products.
func (m *PgModel) CreatePromoRuleTargetProductSet(ctx context.Context, products []*PromoRuleCreateProduct, promoRuleCode, name string, startAt, endAt *time.Time, amount int, typ, target string) (*PromoRuleJoinProductRow, error) {
	contextLogger := log.WithContext(ctx)
	contextLogger.Debugf("postgres: CreatePromoRuleTargetProductSet(ctx, products, promoRuleCode=%q, name=%q, startAt=%v, endAt=%v, amount=%d, typ=%q, target=%q)", promoRuleCode, name, startAt, endAt, amount, typ, target)

	tx, err := m.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, errors.Wrap(err, "postgres: db.BeginTx")
	}
	contextLogger.Debugf("postgres: db begin transaction")

	// TODO: look this up more efficently
	// 1. Create a map of product id to product uuids.
	q1 := "SELECT id, uuid, path, sku, name FROM product"
	rows1, err := tx.QueryContext(ctx, q1)
	if err != nil {
		tx.Rollback()
		return nil, errors.Wrapf(err, "postgres: query context q1=%q", q1)
	}
	defer rows1.Close()

	type product struct {
		id   int
		uuid string
		path string
		sku  string
		name string
	}
	productMap := make(map[string]*product)
	for rows1.Next() {
		var p product
		err = rows1.Scan(&p.id, &p.uuid, &p.path, &p.sku, &p.name)
		if err != nil {
			tx.Rollback()
			return nil, errors.Wrapf(err, "postgres: scan failed")
		}
		productMap[p.uuid] = &p
	}
	if err = rows1.Err(); err != nil {
		return nil, errors.Wrapf(err, "postgres: rows.Err()")
	}

	// check for missing product ids
	// Iterate the create products list passed in to this function.
	// Ensure that each product uuid exists in the maps
	for _, p := range products {
		if _, ok := productMap[p.ProductUUID]; !ok {
			tx.Rollback()
			return nil, ErrProductNotFound
		}
	}

	// 2. Check if the promo rule code exists.
	q2 := "SELECT EXISTS(SELECT 1 FROM promo_rule WHERE promo_rule_code = $1) AS exists"
	var exists bool
	if err := m.db.QueryRowContext(ctx, q2, promoRuleCode).Scan(&exists); err != nil {
		return nil, errors.Wrapf(err, "postgres: tx.QueryRowContext(ctx, q2=%q, promoRuleCode=%q) failed", q2, promoRuleCode)
	}
	if exists {
		contextLogger.Debugf("postgres: promo rule code %q already exists", promoRuleCode)
		return nil, ErrPromoRuleExists
	}
	contextLogger.Debugf("postgres: promo rule code %q is available", promoRuleCode)

	// Create a product set and add the products to it
	q3 := `
		INSERT INTO product_set (created, modified)
		VALUES (NOW(), NOW())
		RETURNING id, uuid
	`
	var productSetID int
	var productSetUUID string
	err = tx.QueryRowContext(ctx, q3).Scan(&productSetID, &productSetUUID)
	if err != nil {
		tx.Rollback()
		return nil, errors.Wrapf(err, "postgres: tx prepare for q3=%q", q3)
	}

	// 4. Insert the products into the product set.
	q4 := `
		INSERT INTO product_set_item
		  (product_set_id, product_id, created, modified)
		VALUES ($1, $2, NOW(), NOW())
		RETURNING
		  id, uuid, product_set_id, product_id, created, modified
	`
	stmt4, err := tx.PrepareContext(ctx, q4)
	if err != nil {
		tx.Rollback()
		return nil, errors.Wrapf(err, "postgres: tx prepare for q4=%q", q4)
	}
	defer stmt4.Close()

	productSet := make([]*ProductSetItemJoinRow, 0, len(products))
	for _, p := range products {
		product := productMap[p.ProductUUID]

		ps := ProductSetItemJoinRow{}
		if err := stmt4.QueryRowContext(ctx, productSetID, product.id).Scan(&ps.id, &ps.UUID, &ps.productSetID, &ps.productID, &ps.Created, &ps.Modified); err != nil {
			tx.Rollback()
			return nil, errors.Wrapf(err, "postgres: stmt3.QueryRowContext(ctx, ...) failed q4=%q", q4)
		}
		// ps.ProductUUID = p.ProductUUID
		ps.ProductPath = product.path

		productSet = append(productSet, &ps)
	}

	// 5. Create the promo rule with the productset as its target.
	q5 := `
		INSERT INTO promo_rule
		(product_set_id, promo_rule_code, name, start_at, end_at, amount, type, target, created, modified)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, NOW(), NOW())
		RETURNING
		id, uuid, product_id, product_set_id, category_id, shipping_tariff_id,
		name, start_at, end_at, amount, total_threshold, type, target, created, modified
`
	r := PromoRuleJoinProductRow{}
	row := tx.QueryRowContext(ctx, q5, productSetID, promoRuleCode, name, startAt, endAt, amount, typ, target)
	if err := row.Scan(&r.id, &r.UUID, &r.productID, &r.productSetID, &r.categoryID, &r.shippingTariffID, &r.Name, &r.StartAt, &r.EndAt, &r.Amount, &r.TotalThreshold, &r.Type, &r.Target, &r.Created, &r.Modified); err != nil {
		return nil, errors.Wrapf(err, "postgres: query row context q5=%q", q5)
	}
	r.ProductSetUUID = &productSetUUID
	// r.ProductUUID = &productUUID
	// r.ProductPath = &productPath
	// r.ProductSKU = &productSKU

	if err = tx.Commit(); err != nil {
		return nil, errors.Wrap(err, "postgres: tx.Commit")
	}

	contextLogger.Debugf("postgres: db commit succeeded")
	return &r, nil
}

// CreatePromoRuleTargetCategory create a promo rule associated to a category.
func (m *PgModel) CreatePromoRuleTargetCategory(ctx context.Context, categoryUUID, promoRuleCode, name string, startAt *time.Time, endAt *time.Time, amount int, typ, target string) (*PromoRuleJoinProductRow, error) {
	contextLogger := log.WithContext(ctx)
	contextLogger.Debugf("postgres: CreatePromoRuleTargetCategory(ctx, categoryUUID=%q, promoRuleCode=%q, name=%q, startAt=%v, endAt=%v, amount=%d, typ=%q, target=%q)", categoryUUID, promoRuleCode, name, startAt, endAt, amount, typ, target)

	tx, err := m.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, errors.Wrap(err, "postgres: db.BeginTx")
	}
	contextLogger.Debugf("postgres: db begin transaction")

	// 1. Check if the category exists.
	q1 := "SELECT id, path FROM category WHERE uuid = $1"
	var categoryID int
	var categoryPath string
	err = tx.QueryRowContext(ctx, q1, categoryUUID).Scan(&categoryID, &categoryPath)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrCategoryNotFound
		}
		tx.Rollback()
		return nil, errors.Wrapf(err, "postgres: tx prepare for q1=%q", q1)
	}
	contextLogger.Debugf("postgres: category (categoryID=%d) exists", categoryID)

	// 2. Check if the promo rule code exists.
	q2 := "SELECT EXISTS(SELECT 1 FROM promo_rule WHERE promo_rule_code = $1) AS exists"
	var exists bool
	if err := m.db.QueryRowContext(ctx, q2, promoRuleCode).Scan(&exists); err != nil {
		return nil, errors.Wrapf(err, "postgres: tx.QueryRowContext(ctx, q2=%q, promoRuleCode=%q) failed", q2, promoRuleCode)
	}
	if exists {
		contextLogger.Debugf("postgres: promo rule code %q already exists", promoRuleCode)
		return nil, ErrPromoRuleExists
	}
	contextLogger.Debugf("postgres: promo rule code %q is available", promoRuleCode)

	// 2. Create the promo rule with the category as its target.
	q3 := `
		INSERT INTO promo_rule
		  (category_id, promo_rule_code, name, start_at, end_at, amount, type, target, created, modified)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, NOW(), NOW())
		RETURNING
		  id, uuid, product_id, promo_rule_code, product_set_id, category_id, shipping_tariff_id,
		  name, start_at, end_at, amount, total_threshold, type, target, created, modified
	`
	r := PromoRuleJoinProductRow{}
	row := m.db.QueryRowContext(ctx, q3, categoryID, promoRuleCode, name, startAt, endAt, amount, typ, target)
	if err := row.Scan(&r.id, &r.UUID, &r.productID, &r.PromoRuleCode, &r.productSetID, &r.categoryID, &r.shippingTariffID, &r.Name, &r.StartAt, &r.EndAt, &r.Amount, &r.TotalThreshold, &r.Type, &r.Target, &r.Created, &r.Modified); err != nil {
		return nil, errors.Wrapf(err, "postgres: query row context q3=%q", q3)
	}
	r.CategoryUUID = &categoryUUID
	r.CategoryPath = &categoryPath

	if err = tx.Commit(); err != nil {
		return nil, errors.Wrap(err, "postgres: tx.Commit")
	}

	contextLogger.Debugf("postgres: db commit succeeded")
	return &r, nil
}

// CreatePromoRuleTargetShippingTariff create a promo rule associated to a shipping tariff.
func (m *PgModel) CreatePromoRuleTargetShippingTariff(ctx context.Context, shippingTariffUUID, promoRuleCode, name string, startAt, endAt *time.Time, amount int, typ, target string) (*PromoRuleJoinProductRow, error) {
	contextLogger := log.WithContext(ctx)
	contextLogger.Debugf("postgres: CreatePromoRuleTargetShippingTariff(ctx, shippingTariffUUID=%q, promoRuleCode=%q, name=%q, startAt=%v, endAt=%v, amount=%d, typ=%q, target=%q)", shippingTariffUUID, promoRuleCode, name, startAt, endAt, amount, typ, target)

	tx, err := m.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, errors.Wrap(err, "postgres: db.BeginTx")
	}
	contextLogger.Debugf("postgres: db begin transaction")

	// 1. Check if the category exists.
	q1 := "SELECT id, shipping_code FROM shipping_tariff WHERE uuid = $1"
	var shippingTariffID int
	var shippingTariffCode string
	err = tx.QueryRowContext(ctx, q1, shippingTariffUUID).Scan(&shippingTariffID, &shippingTariffCode)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrShippingTariffNotFound
		}
		tx.Rollback()
		return nil, errors.Wrapf(err, "postgres: tx prepare for q1=%q", q1)
	}
	contextLogger.Debugf("postgres: shipping tariff id %q found. id=%d, shipping_code=%q", shippingTariffUUID, shippingTariffID, shippingTariffCode)

	// 2. Check if the promo rule code exists.
	q2 := "SELECT EXISTS(SELECT 1 FROM promo_rule WHERE promo_rule_code = $1) AS exists"
	var exists bool
	if err := m.db.QueryRowContext(ctx, q2, promoRuleCode).Scan(&exists); err != nil {
		return nil, errors.Wrapf(err, "postgres: tx.QueryRowContext(ctx, q2=%q, promoRuleCode=%q) failed", q2, promoRuleCode)
	}
	if exists {
		contextLogger.Debugf("postgres: promo rule code %q already exists", promoRuleCode)
		return nil, ErrPromoRuleExists
	}
	contextLogger.Debugf("postgres: promo rule code %q is available", promoRuleCode)

	// 2. Create the promo rule with the shipping_tarif as its target.
	q3 := `
		INSERT INTO promo_rule
		  (shipping_tariff_id, promo_rule_code, name, start_at, end_at, amount, type, target, created, modified)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, NOW(), NOW())
		RETURNING
		  id, uuid, product_id, product_set_id, category_id, shipping_tariff_id,
		  name, start_at, end_at, amount, total_threshold, type, target, created, modified
	`
	r := PromoRuleJoinProductRow{}
	row := m.db.QueryRowContext(ctx, q3, shippingTariffID, promoRuleCode, name, startAt, endAt, amount, typ, target)
	if err := row.Scan(&r.id, &r.UUID, &r.productID, &r.productSetID, &r.categoryID, &r.shippingTariffID, &r.Name, &r.StartAt, &r.EndAt, &r.Amount, &r.TotalThreshold, &r.Type, &r.Target, &r.Created, &r.Modified); err != nil {
		return nil, errors.Wrapf(err, "postgres: query row context q3=%q", q3)
	}
	r.ShippingTariffUUID = &shippingTariffUUID
	r.ShippingTariffCode = &shippingTariffCode

	if err = tx.Commit(); err != nil {
		return nil, errors.Wrap(err, "postgres: tx.Commit")
	}

	contextLogger.Debugf("postgres: db commit succeeded")
	return &r, nil
}

// CreatePromoRuleTargetTotal creates a promo rule with a total threshold.
func (m *PgModel) CreatePromoRuleTargetTotal(ctx context.Context, totalThreshold int, promoRuleCode, name string, startAt, endAt *time.Time, amount int, typ, target string) (*PromoRuleJoinProductRow, error) {
	// 1. Check if the promo rule code exists.
	q1 := "SELECT EXISTS(SELECT 1 FROM promo_rule WHERE promo_rule_code = $1) AS exists"
	var exists bool
	err := m.db.QueryRowContext(ctx, q1, promoRuleCode).Scan(&exists)
	if err != nil {
		return nil, errors.Wrapf(err, "postgres: tx.QueryRowContext(ctx, q1=%q, promoRuleCode=%q) failed", q1, promoRuleCode)
	}
	if exists {
		return nil, ErrPromoRuleExists
	}

	// 2. Create the promo rule with the total as its target.
	q2 := `
		INSERT INTO promo_rule
		  (total_threshold, promo_rule_code, name, start_at, end_at, amount, type, target, created, modified)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, NOW(), NOW())
		RETURNING
		  id, uuid, product_id, product_set_id, category_id, shipping_tariff_id,
		  name, start_at, end_at, amount, total_threshold, type, target, created, modified
	`
	r := PromoRuleJoinProductRow{}
	row := m.db.QueryRowContext(ctx, q2, totalThreshold, promoRuleCode, name, startAt, endAt, amount, typ, target)
	if err := row.Scan(&r.id, &r.UUID, &r.productID, &r.productSetID, &r.categoryID, &r.shippingTariffID, &r.Name, &r.StartAt, &r.EndAt, &r.Amount, &r.TotalThreshold, &r.Type, &r.Target, &r.Created, &r.Modified); err != nil {
		return nil, errors.Wrapf(err, "postgres: query row context q2=%q", q2)
	}
	r.TotalThreshold = &totalThreshold

	return &r, nil
}

// GetPromoRule returns a single promo rule row.
func (m *PgModel) GetPromoRule(ctx context.Context, promoRuleUUID string) (*PromoRuleJoinProductRow, error) {
	q1 := `
		SELECT
		  id, uuid, promo_rule_code, name, start_at, end_at, amount, total_threshold,
		  type, target, created, modified
		FROM promo_rule WHERE uuid = $1
	`
	p := PromoRuleJoinProductRow{}
	row := m.db.QueryRowContext(ctx, q1, promoRuleUUID)
	if err := row.Scan(&p.id, &p.UUID, &p.PromoRuleCode, &p.Name, &p.StartAt, &p.EndAt, &p.Amount, &p.TotalThreshold, &p.Type, &p.Target, &p.Created, &p.Modified); err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrPromoRuleNotFound
		}
		return nil, errors.Wrapf(err, "postgres: query row context scan query=%q promoRuleUUID=%q failed", q1, promoRuleUUID)
	}
	return &p, nil
}

// GetPromoRules returns a list of promo rules.
func (m *PgModel) GetPromoRules(ctx context.Context) ([]*PromoRuleJoinProductRow, error) {
	q1 := `
		SELECT
		  id, uuid, promo_rule_code, name, start_at, end_at, amount, total_threshold, type, target, created, modified
		FROM promo_rule
	`
	rows, err := m.db.QueryContext(ctx, q1)
	if err != nil {
		return nil, errors.Wrapf(err, "postgres: m.db.QueryContext(ctx) failed")
	}
	defer rows.Close()

	rules := make([]*PromoRuleJoinProductRow, 0, 4)
	for rows.Next() {
		var p PromoRuleJoinProductRow
		if err = rows.Scan(&p.id, &p.UUID, &p.PromoRuleCode, &p.Name, &p.StartAt, &p.EndAt, &p.Amount, &p.TotalThreshold, &p.Type, &p.Target, &p.Created, &p.Modified); err != nil {
			if err == sql.ErrNoRows {
				return nil, ErrPromoRuleNotFound
			}
			return nil, errors.Wrap(err, "postgres: scan failed")
		}
		rules = append(rules, &p)
	}
	if err = rows.Err(); err != nil {
		return nil, errors.Wrap(err, "postgres: rows.Err()")
	}
	return rules, nil
}

// DeletePromoRule deletes a promo rule row from the promo_rule table.
func (m *PgModel) DeletePromoRule(ctx context.Context, promoRuleUUID string) error {
	tx, err := m.db.BeginTx(ctx, nil)
	if err != nil {
		return errors.Wrap(err, "postgres: db.BeginTx")
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
		return errors.Wrapf(err, "postgres: query row context failed for q1=%q", q1)
	}

	q2 := "DELETE FROM promo_rule WHERE id = $1"
	_, err = tx.ExecContext(ctx, q2, promoRuleID)
	if err != nil {
		tx.Rollback()
		return errors.Wrapf(err, "postgres: exec context q2=%q", q2)
	}

	if err = tx.Commit(); err != nil {
		return errors.Wrap(err, "postgres: tx.Commit")
	}
	return nil
}
