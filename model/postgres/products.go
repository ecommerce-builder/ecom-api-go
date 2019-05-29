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

// ProductUpdate contains the data required to update an existing product.
type ProductUpdate struct {
	EAN  string
	Path string
	Name string
	Data ProductData
}

// ProductData contains the data stored in the product table data column.
// It provides JSON field tags so the driver can encode and decode.
type ProductData struct {
	Summary string `json:"summary"`
	Desc    string `json:"description"`
	Spec    string `json:"specification"`
}

// Product maps to a products row.
type Product struct {
	id       int
	UUID     string
	SKU      string
	EAN      string
	Path     string
	Name     string
	Data     ProductData
	Created  time.Time
	Modified time.Time
}

// Value marshals ProdutData to a JSON string.
func (pd ProductData) Value() (driver.Value, error) {
	bs, err := json.Marshal(pd)
	if err != nil {
		return nil, errors.Wrap(err, "json marshal failed")
	}
	return string(bs), nil
}

// Scan unmarshals JSON data into a ProductData struct
func (pd *ProductData) Scan(value interface{}) error {
	sv, err := driver.String.ConvertValue(value)
	if err != nil {
		return errors.Wrap(err, "convert value failed")
	}
	if v, ok := sv.([]byte); ok {
		var pdu ProductData
		err := json.Unmarshal(v, &pdu)
		if err != nil {
			return errors.Wrap(err, "json unmarshal failed")
		}
		*pd = pdu
		return nil
	}
	return fmt.Errorf("scan value failed")
}

// ErrProductNotFound is returned by GetProduct when the product with
// the given SKU could not be found in the database.
var ErrProductNotFound = errors.New("model: product not found")

// CreateProduct creates a new product with the given SKU.
func (m *PgModel) CreateProduct(ctx context.Context, sku string, pu *ProductUpdate) (*Product, error) {
	query := `
		INSERT INTO products (
			sku, ean, path, name, data, created, modified
		) VALUES (
			$1, $2, $3, $4, $5, NOW(), NOW()
		) RETURNING
			id, uuid, sku, ean, path, name, data, created, modified
	`
	p := Product{}
	row := m.db.QueryRowContext(ctx, query, sku, pu.EAN, pu.Path, pu.Name, pu.Data)
	if err := row.Scan(&p.id, &p.UUID, &p.SKU, &p.EAN, &p.Path, &p.Name, &p.Data, &p.Created, &p.Modified); err != nil {
		return nil, errors.Wrapf(err, "query scan context sku=%q, query=%q", sku, query)
	}
	return &p, nil
}

// GetProduct returns a single product by SKU.
func (m *PgModel) GetProduct(ctx context.Context, sku string) (*Product, error) {
	query := `
		SELECT id, uuid, sku, ean, path, name, data, created, modified
		FROM products
		WHERE sku = $1
	`
	p := Product{}
	row := m.db.QueryRowContext(ctx, query, sku)
	if err := row.Scan(&p.id, &p.UUID, &p.SKU, &p.EAN, &p.Path, &p.Name, &p.Data, &p.Created, &p.Modified); err != nil {
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
		SELECT id, uuid, sku, ean, path, name, data, created, modified
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
		if err := rows.Scan(&p.id, &p.UUID, &p.SKU, &p.EAN, &p.Path, &p.Name, &p.Data, &p.Created, &p.Modified); err != nil {
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
func (m *PgModel) UpdateProduct(ctx context.Context, sku string, pu *ProductUpdate) (*Product, error) {
	query := `
		UPDATE products
		SET ean = $1, path = $2, name = $3, data = $4, modified = NOW()
		WHERE sku = $5
		RETURNING
			id, uuid, sku, ean, path, name, data, created, modified
	`
	p := Product{}
	row := m.db.QueryRowContext(ctx, query, pu.EAN, pu.Path, pu.Name, pu.Data, sku)
	if err := row.Scan(&p.id, &p.UUID, &p.SKU, &p.EAN, &p.Path, &p.Name, &p.Data, &p.Created, &p.Modified); err != nil {
		return nil, errors.Wrapf(err, "query row context query=%q", query)
	}
	return &p, nil
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
