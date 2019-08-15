package postgres

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/pkg/errors"
)

// ProductPricingEntry contains a tier reference and unit price pair
type ProductPricingEntry struct {
	PricingTierUUID string
	UnitPrice       int
}

// ProductCreate contains the data required to createa a new product.
type ProductCreate struct {
	SKU     string
	EAN     string
	Path    string
	Name    string
	Images  []*CreateImage
	Pricing []*ProductPricingEntry
	Content ProductContent
}

// ProductUpdate contains the data required to update an existing product.
type ProductUpdate struct {
	SKU     string
	EAN     string
	Path    string
	Name    string
	Images  []*CreateImage
	Pricing []*ProductPricingEntry
	Content ProductContent
}

// ProductContent contains the data stored in the product table data column.
// It provides JSON field tags so the driver can encode and decode.
// ProductContent contains the variable JSON data of the product
type ProductContent struct {
	Meta struct {
		Title       string `json:"title"`
		Description string `json:"description"`
	} `json:"meta"`
	Videos        []string `json:"videos"`
	Manuals       []string `json:"manuals"`
	Software      []string `json:"software"`
	Description   string   `json:"description"`
	Specification string   `json:"specification"`
	InTheBox      string   `json:"in_the_box"`
}

// ProductRow maps to a product row.
type ProductRow struct {
	id       int
	UUID     string
	SKU      string
	EAN      string
	Path     string
	Name     string
	Content  ProductContent
	Created  time.Time
	Modified time.Time
}

// ProductFull represents a product row joined with a product_image row
type ProductFull struct {
	id       int
	UUID     string
	SKU      string
	EAN      string
	Path     string
	Name     string
	Images   []*ImageJoinRow
	Pricing  []*ProductPricingRow
	Content  ProductContent
	Created  time.Time
	Modified time.Time
}

// Value marshals ProdutContent to a JSON string.
func (pd ProductContent) Value() (driver.Value, error) {
	bs, err := json.Marshal(pd)
	if err != nil {
		return nil, errors.Wrap(err, "json marshal failed")
	}
	return string(bs), nil
}

// Scan unmarshals JSON data into a ProductContent struct
func (pd *ProductContent) Scan(value interface{}) error {
	sv, err := driver.String.ConvertValue(value)
	if err != nil {
		return errors.Wrap(err, "convert value failed")
	}
	if v, ok := sv.([]byte); ok {
		var content ProductContent
		err := json.Unmarshal(v, &content)
		if err != nil {
			return errors.Wrap(err, "json unmarshal failed")
		}
		*pd = content
		return nil
	}
	return fmt.Errorf("scan value failed")
}

// ErrProductNotFound is returned by GetProduct when the query
// for the product could not be found in the database.
var ErrProductNotFound = errors.New("product not found")

// GetProduct returns a ProductRow by product id.
func (m *PgModel) GetProduct(ctx context.Context, productID string) (*ProductRow, error) {
	query := `
		SELECT
		  id, uuid, sku, ean, path, name, content, created, modified
		FROM
		  product
		WHERE
		  uuid = $1
	`
	p := ProductRow{}
	row := m.db.QueryRowContext(ctx, query, productID)
	if err := row.Scan(&p.id, &p.UUID, &p.SKU, &p.EAN, &p.Path, &p.Name, &p.Content, &p.Created, &p.Modified); err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrProductNotFound
		}
		return nil, errors.Wrapf(err, "query scan context productID=%q query=%q", productID, query)
	}
	return &p, nil
}

// GetProducts returns a list of all products in the product table.
func (m *PgModel) GetProducts(ctx context.Context) ([]*ProductRow, error) {
	query := `
		SELECT
		  id, uuid, sku, ean, path, name, content, created, modified
		FROM
		  product
	`
	rows, err := m.db.QueryContext(ctx, query)
	if err != nil {
		return nil, errors.Wrapf(err, "m.db.QueryContext(ctx) query=%q", query)
	}
	defer rows.Close()

	products := make([]*ProductRow, 0, 256)
	for rows.Next() {
		var p ProductRow
		if err := rows.Scan(&p.id, &p.UUID, &p.SKU, &p.EAN, &p.Path, &p.Name, &p.Content, &p.Created, &p.Modified); err != nil {
			return nil, errors.Wrap(err, "scan failed")
		}
		products = append(products, &p)
	}
	if err = rows.Err(); err != nil {
		return nil, errors.Wrap(err, "rows.Err()")
	}
	return products, nil
}

// ProductsExist accepts a slice of product uuids strings and returns only
// those that can be found in the product table.
func (m *PgModel) ProductsExist(ctx context.Context, productIDs []string) ([]string, error) {
	query := `
		SELECT
		  uuid
		FROM
		  product
		WHERE
		  uuid = ANY($1::UUID[])
	`
	// TODO: sanitise skus
	rows, err := m.db.QueryContext(ctx, query, "{"+strings.Join(productIDs, ",")+"}")
	if err != nil {
		return nil, errors.Wrapf(err, "m.db.QueryContext(ctx,..) query=%q, products=%v", query, productIDs)
	}
	defer rows.Close()

	found := make([]string, 0, 256)
	for rows.Next() {
		var s string
		err = rows.Scan(&s)
		if err != nil {
			return nil, errors.Wrap(err, "scan failed")
		}
		found = append(found, s)
	}
	if err = rows.Err(); err != nil {
		return nil, errors.Wrap(err, "rows.Err()")
	}
	return found, nil
}

// ProductExists return true if there is a row in the product table with
// the given product UUID.
func (m *PgModel) ProductExists(ctx context.Context, productUUID string) (bool, error) {
	query := `SELECT id FROM product WHERE uuid = $1`
	var id int
	err := m.db.QueryRowContext(ctx, query, productUUID).Scan(&id)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, ErrProductNotFound
		}
		return false, errors.Wrapf(err, "query row context productUUID=%q query=%q", productUUID, query)
	}
	return true, nil
}

// ProductExistsBySKU return true if there is a row in the product table with
// the given SKU.
func (m *PgModel) ProductExistsBySKU(ctx context.Context, sku string) (bool, error) {
	query := `SELECT id FROM product WHERE sku = $1`
	var id int
	err := m.db.QueryRowContext(ctx, query, sku).Scan(&id)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		return false, errors.Wrapf(err, "query row context sku=%q query=%q", sku, query)
	}
	return true, nil
}

// CreateProduct updates the details of a product with the given product id.
func (m *PgModel) CreateProduct(ctx context.Context, pu *ProductCreate) (*ProductFull, error) {

	fmt.Printf("%#v\n", pu)
	fmt.Println("-------------")

	tx, err := m.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, errors.Wrap(err, "db.BeginTx")
	}

	q1 := `
		INSERT INTO product
		  (sku, ean, path, name, content, created, modified)
		VALUES
		  ($1, $2, $3, $4, $5, NOW(), NOW())
		RETURNING
		  id, uuid, sku, ean, path, name, content, created, modified
	`
	p := ProductRow{}
	row := tx.QueryRowContext(ctx, q1, pu.SKU, pu.EAN, pu.Path, pu.Name, pu.Content)
	if err := row.Scan(&p.id, &p.UUID, &p.SKU, &p.EAN, &p.Path, &p.Name, &p.Content, &p.Created, &p.Modified); err != nil {
		tx.Rollback()
		return nil, errors.Wrapf(err, "query row context query=%q", q1)
	}

	// Delete all existing products. This is not the most efficient
	// way, but is easier that comparing the state of a list of new
	// images with the underlying database. Product updates don't
	// effect the customer's experience.
	q2 := `DELETE FROM product_image WHERE product_id = $1`
	if _, err = tx.ExecContext(ctx, q2, p.id); err != nil {
		tx.Rollback()
		return nil, errors.Wrapf(err, "model: delete product_image query=%q", q2)
	}

	// The inner sub select uses a pri of 10 if now rows exist for the given product
	// or pri + 10 for each subseqent row.
	q3 := `
		INSERT INTO product_image (
			product_id,
			w, h, path, typ,
			ori, up,
			pri, size, q,
			gsurl, created, modified
		) VALUES (
			$1,
			$2, $3, $4, $5,
			$6, false,
			(
				SELECT
					CASE WHEN COUNT(1) > 0
					THEN MAX(pri)+10
					ELSE 10
				END
				AS pri
				FROM product_image
				WHERE id = $7
			), $8, $9,
			$10, NOW(), NOW()
		) RETURNING
		  id, uuid, product_id, w,
		  h, path, typ, ori, up, pri, size, q,
		  gsurl, data, created, modified
	`
	stmt3, err := tx.PrepareContext(ctx, q3)
	if err != nil {
		tx.Rollback()
		return nil, errors.Wrapf(err, "tx prepare for query=%q", q3)
	}
	defer stmt3.Close()

	images := make([]*ImageJoinRow, 0)
	for _, img := range pu.Images {
		var pi ImageJoinRow
		row := stmt3.QueryRowContext(ctx, p.id,
			img.W, img.H, img.Path, img.Typ,
			img.Ori,
			p.id, img.Size, img.Q,
			img.GSURL)
		if err := row.Scan(&pi.id, &pi.UUID, &pi.productID, &pi.W,
			&pi.H, &pi.Path, &pi.Typ, &pi.Ori, &pi.Up, &pi.Pri, &pi.Size, &pi.Q,
			&pi.GSURL, &pi.Data, &p.Created, &p.Modified); err != nil {
			tx.Rollback()
			return nil, errors.Wrapf(err, "row.Scan failed")
		}
		pi.ProductUUID = p.UUID
		images = append(images, &pi)
	}

	// insert or update product pricing matrix (uses and 'upsert' with ON CONFLICT)
	q4 := `
		SELECT id
		FROM pricing_tier
		WHERE uuid = $1
	`
	stmt4, err := tx.PrepareContext(ctx, q4)
	if err != nil {
		tx.Rollback()
		return nil, errors.Wrapf(err, "tx prepare for query=%q", q4)
	}
	defer stmt4.Close()

	q5 := `
		INSERT INTO product_pricing
		  (pricing_tier_id, product_id, unit_price, created, modified)
		VALUES
		  ($1, $2, $3, NOW(), NOW())
		ON CONFLICT (pricing_tier_id, product_id)
		DO UPDATE
		  SET
		    unit_price = $4, modified = NOW()
		  WHERE
		    excluded.product_id = $5 AND excluded.pricing_tier_id = $6
		RETURNING
		  id, uuid, pricing_tier_id, product_id, unit_price, created, modified
	`
	stmt5, err := tx.PrepareContext(ctx, q5)
	if err != nil {
		tx.Rollback()
		return nil, errors.Wrapf(err, "tx prepare for query=%q", q5)
	}
	defer stmt5.Close()

	pricing := make([]*ProductPricingRow, 0, 4)
	for _, r := range pu.Pricing {
		var pricingTierID int
		if err := stmt4.QueryRowContext(ctx, r.PricingTierUUID).Scan(&pricingTierID); err != nil {
			if err == sql.ErrNoRows {
				tx.Rollback()
				return nil, ErrPricingTierNotFound
			}
			tx.Rollback()
		}

		row = stmt5.QueryRowContext(ctx, pricingTierID, p.id, r.UnitPrice,
			r.UnitPrice, p.id, pricingTierID)
		var pp ProductPricingRow
		if err := row.Scan(&pp.id, &pp.UUID, &pp.pricingTierID, &pp.productID, &pp.UnitPrice, &pp.Created, &pp.Modified); err != nil {
			return nil, errors.Wrap(err, "scan failed")
		}
		pricing = append(pricing, &pp)
	}
	productFull := ProductFull{
		id:       p.id,
		UUID:     p.UUID,
		SKU:      p.SKU,
		EAN:      p.EAN,
		Path:     p.Path,
		Name:     p.Name,
		Images:   images,
		Pricing:  pricing,
		Content:  p.Content,
		Created:  p.Created,
		Modified: p.Modified,
	}
	if err = tx.Commit(); err != nil {
		return nil, errors.Wrap(err, "tx.Commit")
	}
	return &productFull, nil
}

// UpdateProduct updates the details of a product with the given product id.
func (m *PgModel) UpdateProduct(ctx context.Context, productID string, pu *ProductUpdate) (*ProductFull, error) {

	fmt.Printf("%#v\n", pu)
	fmt.Println("-------------")
	tx, err := m.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, errors.Wrap(err, "db.BeginTx")
	}

	q1 := `
		UPDATE product
		SET
		  ean = $1, path = $2, name = $3, content = $4, modified = NOW()
		WHERE
		  uuid = $5
		RETURNING
		  id, uuid, sku, ean, path, name, content, created, modified`
	row := tx.QueryRowContext(ctx, q1, pu.EAN, pu.Path, pu.Name, pu.Content, productID)

	p := ProductRow{}
	if err := row.Scan(&p.id, &p.UUID, &p.SKU, &p.EAN, &p.Path, &p.Name, &p.Content, &p.Created, &p.Modified); err != nil {
		tx.Rollback()
		return nil, errors.Wrapf(err, "query row context query=%q", q1)
	}

	// Delete all existing products. This is not the most efficient
	// way, but is easier that comparing the state of a list of new
	// images with the underlying database. Product updates don't
	// effect the customer's experience.
	q2 := `DELETE FROM product_image WHERE product_id = $1`
	if _, err = tx.ExecContext(ctx, q2, p.id); err != nil {
		tx.Rollback()
		return nil, errors.Wrapf(err, "model: delete product_image query=%q", q2)
	}

	// The inner sub select uses a pri of 10 if now rows exist for the given product
	// or pri + 10 for each subseqent row.
	q3 := `
		INSERT INTO product_image (
			product_id,
			w, h, path, typ,
			ori, up,
			pri, size, q,
			gsurl, created, modified
		) VALUES (
			$1,
			$2, $3, $4, $5,
			$6, false,
			(
				SELECT
					CASE WHEN COUNT(1) > 0
					THEN MAX(pri)+10
					ELSE 10
				END
				AS pri
				FROM product_image
				WHERE id = $7
			), $8, $9,
			$10, NOW(), NOW()
		) RETURNING
		  id, product_id, uuid, w,
		  h, path, typ, ori, up, pri, size, q,
		  gsurl, data, created, modified
	`
	stmt3, err := tx.PrepareContext(ctx, q3)
	if err != nil {
		tx.Rollback()
		return nil, errors.Wrapf(err, "tx prepare for query=%q", q3)
	}
	defer stmt3.Close()

	images := make([]*ImageJoinRow, 0)
	for _, img := range pu.Images {
		var pi ImageJoinRow
		row := stmt3.QueryRowContext(ctx, p.id,
			img.W, img.H, img.Path, img.Typ,
			img.Ori,
			p.id, img.Size, img.Q,
			img.GSURL)
		if err := row.Scan(&pi.id, &pi.productID, &pi.UUID, &pi.W,
			&pi.H, &pi.Path, &pi.Typ, &pi.Ori, &pi.Up, &pi.Pri, &pi.Size, &pi.Q,
			&pi.GSURL, &pi.Data, &p.Created, &p.Modified); err != nil {
			tx.Rollback()
			return nil, errors.Wrapf(err, "row.Scan failed")
		}
		images = append(images, &pi)
	}

	// insert or update product pricing matrix (uses and 'upsert' with ON CONFLICT)
	q4 := `
		SELECT id
		FROM pricing_tier
		WHERE uuid = $1
	`
	stmt4, err := tx.PrepareContext(ctx, q4)
	if err != nil {
		tx.Rollback()
		return nil, errors.Wrapf(err, "tx prepare for query=%q", q4)
	}
	defer stmt4.Close()

	q5 := `
		INSERT INTO product_pricing
		  (pricing_tier_id, product_id, unit_price, created, modified)
		VALUES
		  ($1, $2, $3, NOW(), NOW())
		ON CONFLICT (pricing_tier_id, product_id)
		DO UPDATE
		  SET
		    unit_price = $4, modified = NOW()
		  WHERE
		    excluded.product_id = $5 AND excluded.pricing_tier_id = $6
		RETURNING
		  id, uuid, pricing_tier_id, product_id, unit_price, created, modified
	`
	stmt5, err := tx.PrepareContext(ctx, q5)
	if err != nil {
		tx.Rollback()
		return nil, errors.Wrapf(err, "tx prepare for query=%q", q5)
	}
	defer stmt5.Close()

	pricing := make([]*ProductPricingRow, 0, 4)
	for _, r := range pu.Pricing {
		var pricingTierID int
		if err := stmt4.QueryRowContext(ctx, r.PricingTierUUID).Scan(&pricingTierID); err != nil {
			if err == sql.ErrNoRows {
				tx.Rollback()
				return nil, ErrPricingTierNotFound
			}
			tx.Rollback()
		}

		row = stmt5.QueryRowContext(ctx, pricingTierID, p.id, r.UnitPrice,
			r.UnitPrice, p.id, pricingTierID)
		var pp ProductPricingRow
		if err := row.Scan(&pp.id, &pp.UUID, &pp.pricingTierID, &pp.productID, &pp.UnitPrice, &pp.Created, &pp.Modified); err != nil {
			return nil, errors.Wrap(err, "scan failed")
		}
		pricing = append(pricing, &pp)
	}
	productFull := ProductFull{
		id:       p.id,
		UUID:     p.UUID,
		SKU:      p.SKU,
		EAN:      p.EAN,
		Path:     p.Path,
		Name:     p.Name,
		Images:   images,
		Pricing:  pricing,
		Content:  p.Content,
		Created:  p.Created,
		Modified: p.Modified,
	}
	if err = tx.Commit(); err != nil {
		return nil, errors.Wrap(err, "tx.Commit")
	}
	return &productFull, nil
}

// DeleteProduct delete the product with the given UUID.
func (m *PgModel) DeleteProduct(ctx context.Context, productUUID string) error {
	tx, err := m.db.BeginTx(ctx, nil)
	if err != nil {
		return errors.Wrap(err, "db.BeginTx")
	}

	query := `SELECT id FROM product WHERE uuid = $1`
	var productID int
	err = tx.QueryRowContext(ctx, query, productUUID).Scan(&productID)
	if err != nil {
		if err == sql.ErrNoRows {
			tx.Rollback()
			return ErrProductNotFound
		}
		tx.Rollback()
		return errors.Wrapf(err, "query row context failed for query=%q", query)
	}

	query = "DELETE FROM product_pricing WHERE product_id = $1"
	_, err = tx.ExecContext(ctx, query, productID)
	if err != nil {
		tx.Rollback()
		return errors.Wrapf(err, "exec context query=%q", query)
	}

	query = "DELETE FROM category_product WHERE product_id = $1"
	_, err = tx.ExecContext(ctx, query, productID)
	if err != nil {
		tx.Rollback()
		return errors.Wrapf(err, "exec context query=%q", query)
	}

	query = "DELETE FROM product WHERE id = $1"
	_, err = tx.ExecContext(ctx, query, productID)
	if err != nil {
		tx.Rollback()
		return errors.Wrapf(err, "exec context query=%q", query)
	}

	if err = tx.Commit(); err != nil {
		return errors.Wrap(err, "tx.Commit")
	}
	return nil
}
