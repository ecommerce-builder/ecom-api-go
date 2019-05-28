package postgres

import (
	"context"
	"database/sql"
	"time"

	"github.com/pkg/errors"
)

// CreateProductImage struct contains the data required to store a new product image.
type CreateProductImage struct {
	SKU   string
	W     uint
	H     uint
	Path  string
	Typ   string
	Ori   bool
	Pri   uint
	Size  uint
	Q     uint
	GSURL string
	Data  interface{}
}

// ProductImage struct holds a row of the product_images table.
type ProductImage struct {
	id        uint
	ProductID uint
	UUID      string
	SKU       string
	W         uint
	H         uint
	Path      string
	Typ       string
	Ori       bool
	Up        bool
	Pri       uint
	Size      uint
	Q         uint
	GSURL     string
	Data      interface{}
	Created   time.Time
	Modified  time.Time
}

// CreateImageEntry writes a new image entry to the product_images table.
func (m *PgModel) CreateImageEntry(ctx context.Context, c *CreateProductImage) (*ProductImage, error) {
	query := `
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
			$8, $9, $10,
			$11, NOW(), NOW()
		) RETURNING
			id, product_id, uuid, sku, w, h, path, typ, ori, up, pri, size, q,
			gsurl, data, created, modified
	`
	p := ProductImage{}
	err := m.db.QueryRowContext(ctx, query, c.SKU, c.SKU,
		c.W, c.H, c.Path, c.Typ,
		c.Ori,
		c.Pri, c.Size, c.Q,
		c.GSURL).Scan(&p.id, &p.ProductID, &p.UUID, &p.SKU, &p.W, &p.H, &p.Path, &p.Typ, &p.Ori, &p.Up, &p.Pri, &p.Size, &p.Q, &p.GSURL, &p.Data, &p.Created, &p.Modified)
	if err != nil {
		return nil, err
	}

	return &p, nil
}

// GetImagesBySKU returns a slice of Images associated to a given product SKU.
func (m *PgModel) GetImagesBySKU(ctx context.Context, sku string) ([]*ProductImage, error) {
	query := `
		SELECT
		  id, product_id, uuid, sku, w, h, path, typ, ori, up, pri, size, q,
		  gsurl, data, created, modified
		FROM product_images
		WHERE sku = $1
		ORDER BY pri ASC
	`
	rows, err := m.db.QueryContext(ctx, query, sku)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	images := make([]*ProductImage, 0, 16)
	for rows.Next() {
		p := ProductImage{}
		err = rows.Scan(&p.id, &p.ProductID, &p.UUID, &p.SKU, &p.W, &p.H, &p.Path, &p.Typ, &p.Ori, &p.Up, &p.Pri, &p.Size, &p.Q, &p.GSURL, &p.Data, &p.Created, &p.Modified)
		if err != nil {
			return nil, err
		}
		images = append(images, &p)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return images, nil
}

// ImagePathExists return true if the image with the given path exists in
// the database.
func (m *PgModel) ImagePathExists(ctx context.Context, path string) (bool, error) {
	query := `SELECT id FROM product_images WHERE path = $1`
	var id int
	if err := m.db.QueryRowContext(ctx, query, path).Scan(&id); err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		return false, errors.Wrapf(err, "query row context path=%q query=%q", path, query)
	}
	return true, nil
}

// GetProductImageByUUID returns a ProductImage by the given UUID.
// If the product is not found returns nil, error indicating not found.
func (m *PgModel) GetProductImageByUUID(ctx context.Context, uuid string) (*ProductImage, error) {
	query := `
		SELECT id, product_id, uuid, sku, w, h, path, typ, ori, up, pri, size, q,
		gsurl, data, created, modified
		FROM product_images
		WHERE uuid = $1
	`
	p := ProductImage{}
	if err := m.db.QueryRowContext(ctx, query, uuid).Scan(&p.id, &p.ProductID, &p.UUID, &p.SKU, &p.W, &p.H, &p.Path, &p.Typ, &p.Ori, &p.Up, &p.Pri, &p.Size, &p.Q, &p.GSURL, &p.Data, &p.Created, &p.Modified); err != nil {
		return nil, err
	}
	return &p, nil
}

// ImageUUIDExists return true if the image with the given UUID exists in
// the database.
func (m *PgModel) ImageUUIDExists(ctx context.Context, uuid string) (bool, error) {
	query := `SELECT id FROM product_images WHERE uuid = $1`
	var id int
	if err := m.db.QueryRowContext(ctx, query, uuid).Scan(&id); err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		return false, errors.Wrapf(err, "query row context uuid=%q query=%q", uuid, query)
	}
	return true, nil
}

// ConfirmImageUploaded updates the `up` column to true to indicate the
// uploaded has taken place.
func (m *PgModel) ConfirmImageUploaded(ctx context.Context, uuid string) (*ProductImage, error) {
	query := `
		UPDATE product_images
		SET up = 't', modified = NOW()
		WHERE uuid = $1
		RETURNING id, product_id, uuid, sku, w, h, path, typ, ori, up, pri, size, q,
		gsurl, data, created, modified
	`
	p := ProductImage{}
	err := m.db.QueryRowContext(ctx, query, uuid).Scan(&p.id, &p.ProductID, &p.UUID, &p.SKU, &p.W, &p.H, &p.Path, &p.Typ, &p.Ori, &p.Up, &p.Pri, &p.Size, &p.Q, &p.GSURL, &p.Data, &p.Created, &p.Modified)
	if err != nil {
		return nil, err
	}
	return &p, nil
}

// DeleteProductImage deletes an image entry row from the product_images
// table by UUID.
func (m *PgModel) DeleteProductImage(ctx context.Context, uuid string) (int64, error) {
	query := `
		DELETE FROM product_images
		WHERE uuid = $1
	`
	res, err := m.db.ExecContext(ctx, query, uuid)
	if err != nil {
		return -1, err
	}
	count, err := res.RowsAffected()
	if err != nil {
		return -1, err
	}
	return count, nil
}

// DeleteAllProductImages Images deletes all images from the product_images table
//  associated to the product with the given SKU.
func (m *PgModel) DeleteAllProductImages(ctx context.Context, sku string) (int64, error) {
	query := `
		DELETE FROM product_images
		WHERE sku = $1
	`
	res, err := m.db.ExecContext(ctx, query, sku)
	if err != nil {
		return -1, err
	}
	count, err := res.RowsAffected()
	if err != nil {
		return -1, err
	}
	return count, nil
}
