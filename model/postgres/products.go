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
	TierRef   string
	UnitPrice float64
}

// ProductCreateUpdate contains the data required to update an existing product.
type ProductCreateUpdate struct {
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

// Product maps to a products row.
type Product struct {
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

// ProductFull maps to a products row joined to product_images rows
type ProductFull struct {
	id       int
	UUID     string
	SKU      string
	EAN      string
	Path     string
	Name     string
	Images   []*Image
	Pricing  []*ProductPricing
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

// ErrProductNotFound is returned by GetProduct when the product with
// the given SKU could not be found in the database.
var ErrProductNotFound = errors.New("model: product not found")

// CreateProduct creates a new product with the given SKU.
func (m *PgModel) CreateProduct(ctx context.Context, sku, ean, path, name string, content ProductContent) (*Product, error) {
	query := `
		INSERT INTO products (
			sku, ean, path, name, content, created, modified
		) VALUES (
			$1, $2, $3, $4, $5, NOW(), NOW()
		) RETURNING
			id, uuid, sku, ean, path, name, content, created, modified
	`
	p := Product{}
	row := m.db.QueryRowContext(ctx, query, sku, ean, path, name, content)
	if err := row.Scan(&p.id, &p.UUID, &p.SKU, &p.EAN, &p.Path, &p.Name, &p.Content, &p.Created, &p.Modified); err != nil {
		return nil, errors.Wrapf(err, "query scan context sku=%q, query=%q", sku, query)
	}
	return &p, nil
}

// GetProduct returns a single product by SKU.
func (m *PgModel) GetProduct(ctx context.Context, sku string) (*Product, error) {
	query := `
		SELECT id, uuid, sku, ean, path, name, content, created, modified
		FROM products
		WHERE sku = $1
	`
	p := Product{}
	row := m.db.QueryRowContext(ctx, query, sku)
	if err := row.Scan(&p.id, &p.UUID, &p.SKU, &p.EAN, &p.Path, &p.Name, &p.Content, &p.Created, &p.Modified); err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrProductNotFound
		}
		return nil, errors.Wrapf(err, "query scan context sku=%q query=%q", sku, query)
	}
	return &p, nil
}

// GetProducts returns a list of all products in the products table.
func (m *PgModel) GetProducts(ctx context.Context) ([]*Product, error) {
	query := `
		SELECT id, uuid, sku, ean, path, name, content, created, modified
		FROM products
	`
	rows, err := m.db.QueryContext(ctx, query)
	if err != nil {
		return nil, errors.Wrapf(err, "m.db.QueryContext(ctx) query=%q", query)
	}
	defer rows.Close()

	products := make([]*Product, 0, 256)
	for rows.Next() {
		var p Product
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

// ProductsExist accepts a slice of product SKU strings and returns only
// those that can be found in the products table.
func (m *PgModel) ProductsExist(ctx context.Context, skus []string) ([]string, error) {
	query := `
		SELECT sku FROM products
		WHERE sku = ANY($1::varchar[])
	`
	// TODO: sanitise skus
	rows, err := m.db.QueryContext(ctx, query, "{"+strings.Join(skus, ",")+"}")
	if err != nil {
		return nil, errors.Wrapf(err, "m.db.QueryContext(ctx,..) query=%q, skus=%v", query, skus)
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

// ProductExists return true if there is a row in the products table with
// the given SKU.
func (m *PgModel) ProductExists(ctx context.Context, sku string) (bool, error) {
	query := `SELECT id FROM products WHERE sku = $1`
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

// UpdateProduct updates the details of a product with the given SKU.
func (m *PgModel) UpdateProduct(ctx context.Context, sku string, pu *ProductCreateUpdate) (*ProductFull, error) {
	tx, err := m.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, errors.Wrap(err, "db.BeginTx")
	}
	query := `
		INSERT INTO products (sku, ean, path, name, content, created, modified)
		VALUES ($1, $2, $3, $4, $5, NOW(), NOW())
		ON CONFLICT (sku)
		DO UPDATE
		  SET
		    ean = $6, path = $7, name = $8, content = $9, modified = NOW()
		  WHERE excluded.sku = $10
		RETURNING
		  id, uuid, sku, ean, path, name, content, created, modified
	`
	p := Product{}
	row := m.db.QueryRowContext(ctx, query, sku, pu.EAN, pu.Path, pu.Name, pu.Content,
		pu.EAN, pu.Path, pu.Name, pu.Content, sku)
	if err := row.Scan(&p.id, &p.UUID, &p.SKU, &p.EAN, &p.Path, &p.Name, &p.Content, &p.Created, &p.Modified); err != nil {
		tx.Rollback()
		return nil, errors.Wrapf(err, "query row context query=%q", query)
	}

	// Delete all existing products. This is not the most efficient
	// way, but is easier that comparing the state of a list of new
	// images with the underlying database. Product updates don't
	// effect the customer's experience.
	query = `DELETE FROM product_images WHERE sku=$1`
	if _, err = tx.ExecContext(ctx, query, sku); err != nil {
		tx.Rollback()
		return nil, errors.Wrapf(err, "model: delete product_images query=%q", query)
	}

	// The inner sub select uses a pri of 10 if now rows exist for the given product
	// or pri + 10 for each subseqent row.
	query = `
		INSERT INTO product_images (
			product_id, sku,
			w, h, path, typ,
			ori, up,
			pri, size, q,
			gsurl, created, modified
		) VALUES (
			(SELECT id FROM products WHERE sku = $1), $2,
			$3, $4, $5, $6,
			$7, false,
			(
				SELECT
					CASE WHEN COUNT(1) > 0
					THEN MAX(pri)+10
					ELSE 10
				END
				AS pri
				FROM product_images
				WHERE sku=$8
			), $9, $10,
			$11, NOW(), NOW()
		) RETURNING
		  id, product_id, uuid, sku, w,
		  h, path, typ, ori, up, pri, size, q,
		  gsurl, data, created, modified
	`
	stmt1, err := tx.PrepareContext(ctx, query)
	if err != nil {
		tx.Rollback()
		return nil, errors.Wrapf(err, "tx prepare for query=%q", query)
	}
	defer stmt1.Close()

	images := make([]*Image, 0)
	for _, img := range pu.Images {
		var pi Image
		row := stmt1.QueryRowContext(ctx, sku, sku,
			img.W, img.H, img.Path, img.Typ,
			img.Ori,
			sku, img.Size, img.Q,
			img.GSURL)
		if err := row.Scan(&pi.id, &pi.ProductID, &pi.UUID, &pi.SKU, &pi.W,
			&pi.H, &pi.Path, &pi.Typ, &pi.Ori, &pi.Up, &pi.Pri, &pi.Size, &pi.Q,
			&pi.GSURL, &pi.Data, &p.Created, &p.Modified); err != nil {
			tx.Rollback()
			return nil, errors.Wrapf(err, "row.Scan failed")
		}
		images = append(images, &pi)
	}

	// insert or update product pricing matrix (uses and 'upsert' with ON CONFLICT)
	query = `
		INSERT INTO product_pricing (tier_ref, sku, unit_price, created, modified)
		VALUES ($1, $2, $3, NOW(), NOW())
		ON CONFLICT (tier_ref, sku)
		DO UPDATE
		  SET
		    unit_price = $4, modified = NOW()
		  WHERE
		    excluded.sku = $5 AND excluded.tier_ref = $6
		RETURNING
		  id, tier_ref, sku, unit_price, created, modified
	`
	stmt2, err := tx.PrepareContext(ctx, query)
	if err != nil {
		tx.Rollback()
		return nil, errors.Wrapf(err, "tx prepare for query=%q", query)
	}
	defer stmt1.Close()

	pricing := make([]*ProductPricing, 0, 4)
	for _, price := range pu.Pricing {
		row = stmt2.QueryRowContext(ctx, price.TierRef, sku, price.UnitPrice, price.UnitPrice, sku, price.TierRef)
		var pp ProductPricing
		if err := row.Scan(&pp.id, &pp.TierRef, &pp.SKU, &pp.UnitPrice, &pp.Created, &pp.Modified); err != nil {
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

// DeleteProduct delete the product with the given SKU returning the nu
func (m *PgModel) DeleteProduct(ctx context.Context, sku string) error {
	query := `
		DELETE FROM products
		WHERE sku = $1
	`
	_, err := m.db.ExecContext(ctx, query, sku)
	if err != nil {
		return errors.Wrapf(err, "exec context query=%q", query)
	}
	return nil
}