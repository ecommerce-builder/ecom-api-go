package postgres

import (
	"context"
	"database/sql"
	"time"

	"github.com/pkg/errors"
)

// ErrImageNotFound is returned when any query
// for an image has no results in the resultset.
var ErrImageNotFound = errors.New("image not found")

// CreateImage struct contains the data required to store a new product image.
type CreateImage struct {
	W     int
	H     int
	Path  string
	Typ   string
	Ori   bool
	Pri   int
	Size  int
	Q     int
	GSURL string
	Data  interface{}
}

// ImageRow struct holds a row of the product_image table.
type ImageRow struct {
	id        int
	UUID      string
	productID int
	W         int
	H         int
	Path      string
	Typ       string
	Ori       bool
	Up        bool
	Pri       int
	Size      int
	Q         int
	GSURL     string
	Data      interface{}
	Created   time.Time
	Modified  time.Time
}

// ImageJoinRow struct holds a row of the product_image table.
type ImageJoinRow struct {
	id          int
	UUID        string
	productID   int
	ProductUUID string
	W           int
	H           int
	Path        string
	Typ         string
	Ori         bool
	Up          bool
	Pri         int
	Size        int
	Q           int
	GSURL       string
	Data        interface{}
	Created     time.Time
	Modified    time.Time
}

// CreateImageEntry writes a new image entry to the product_image table.
func (m *PgModel) CreateImageEntry(ctx context.Context, productUUID string, c *CreateImage) (*ImageJoinRow, error) {
	tx, err := m.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, errors.Wrap(err, "db.BeginTx")
	}

	q1 := "SELECT id FROM product WHERE uuid = $1"
	var productID int
	err = tx.QueryRowContext(ctx, q1, productUUID).Scan(&productID)
	if err != nil {
		if err == sql.ErrNoRows {
			tx.Rollback()
			return nil, ErrProductNotFound
		}
		tx.Rollback()
		return nil, errors.Wrapf(err, "query row context failed for q1=%q", q1)
	}

	q2 := `
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
			$7, $8, $9,
			$10, NOW(), NOW()
		) RETURNING
			id, uuid, product_id, (SELECT uuid FROM product WHERE id = $11), w, h, path, typ, ori, up, pri, size, q,
			gsurl, data, created, modified
	`
	p := ImageJoinRow{}
	err = tx.QueryRowContext(ctx, q2, productID,
		c.W, c.H, c.Path, c.Typ,
		c.Ori,
		c.Pri, c.Size, c.Q,
		c.GSURL, productID).Scan(&p.id, &p.UUID, &p.productID, &p.ProductUUID, &p.W, &p.H, &p.Path, &p.Typ, &p.Ori, &p.Up, &p.Pri, &p.Size, &p.Q, &p.GSURL, &p.Data, &p.Created, &p.Modified)
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	if err = tx.Commit(); err != nil {
		return nil, errors.Wrap(err, "tx.Commit")
	}
	return &p, nil
}

// GetImages returns a slice of Images associated to a given product SKU.
func (m *PgModel) GetImages(ctx context.Context, productUUID string) ([]*ImageJoinRow, error) {
	tx, err := m.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, errors.Wrap(err, "postgres: db.BeginTx")
	}

	q1 := "SELECT id FROM product WHERE uuid = $1"
	var productID int
	err = tx.QueryRowContext(ctx, q1, productUUID).Scan(&productID)
	if err != nil {
		if err == sql.ErrNoRows {
			tx.Rollback()
			return nil, ErrProductNotFound
		}
		tx.Rollback()
		return nil, errors.Wrapf(err, "postgres: query row context failed for q1=%q", q1)
	}

	q2 := `
		SELECT
		  i.id, i.uuid, product_id, p.uuid as product_uuid, w, h, i.path, typ, ori, up, pri, size, q,
		  gsurl, data, i.created, i.modified
		FROM product_image AS i
		INNER JOIN product AS p
		  ON p.id = i.product_id
		WHERE product_id = $1
		ORDER BY i.pri ASC
	`
	rows, err := tx.QueryContext(ctx, q2, productID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	images := make([]*ImageJoinRow, 0, 4)
	for rows.Next() {
		i := ImageJoinRow{}
		err = rows.Scan(&i.id, &i.UUID, &i.productID, &i.ProductUUID, &i.W, &i.H, &i.Path, &i.Typ, &i.Ori, &i.Up, &i.Pri, &i.Size, &i.Q, &i.GSURL, &i.Data, &i.Created, &i.Modified)
		if err != nil {
			return nil, err
		}
		images = append(images, &i)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}

	if err = tx.Commit(); err != nil {
		return nil, errors.Wrap(err, "postgres: tx.Commit failed")
	}
	return images, nil
}

// GetImagesBySKU returns a slice of Images associated to a given product SKU.
func (m *PgModel) GetImagesBySKU(ctx context.Context, sku string) ([]*ImageRow, error) {
	query := `
		SELECT
		  id, product_id, uuid, sku, w, h, path, typ, ori, up, pri, size, q,
		  gsurl, data, created, modified
		FROM product_image
		WHERE sku = $1
		ORDER BY pri ASC
	`
	rows, err := m.db.QueryContext(ctx, query, sku)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	images := make([]*ImageRow, 0, 16)
	for rows.Next() {
		p := ImageRow{}
		err = rows.Scan(&p.id, &p.productID, &p.UUID, &p.W, &p.H, &p.Path, &p.Typ, &p.Ori, &p.Up, &p.Pri, &p.Size, &p.Q, &p.GSURL, &p.Data, &p.Created, &p.Modified)
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
	query := `SELECT id FROM product_image WHERE path = $1`
	var id int
	if err := m.db.QueryRowContext(ctx, query, path).Scan(&id); err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		return false, errors.Wrapf(err, "query row context path=%q query=%q", path, query)
	}
	return true, nil
}

// GetProductImage returns a ImageJoinRow for the given image uuid.
func (m *PgModel) GetProductImage(ctx context.Context, imageUUID string) (*ImageJoinRow, error) {
	query := `
		SELECT
		  i.id, i.uuid, product_id, p.uuid as product_uuid, w, h, i.path, typ, ori, up, pri, size, q,
		  gsurl, data, i.created, i.modified
		FROM product_image AS i
		INNER JOIN product AS p
		  ON p.id = i.product_id
		WHERE i.uuid = $1
	`
	p := ImageJoinRow{}
	if err := m.db.QueryRowContext(ctx, query, imageUUID).Scan(&p.id, &p.UUID, &p.productID, &p.ProductUUID, &p.W, &p.H, &p.Path, &p.Typ, &p.Ori, &p.Up, &p.Pri, &p.Size, &p.Q, &p.GSURL, &p.Data, &p.Created, &p.Modified); err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrImageNotFound
		}
		return nil, err
	}
	return &p, nil
}

// ImageUUIDExists return true if the image with the given UUID exists in
// the database.
func (m *PgModel) ImageUUIDExists(ctx context.Context, uuid string) (bool, error) {
	query := `SELECT id FROM product_image WHERE uuid = $1`
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
func (m *PgModel) ConfirmImageUploaded(ctx context.Context, uuid string) (*ImageRow, error) {
	query := `
		UPDATE product_image
		SET up = 't', modified = NOW()
		WHERE uuid = $1
		RETURNING id, product_id, uuid, sku, w, h, path, typ, ori, up, pri, size, q,
		gsurl, data, created, modified
	`
	p := ImageRow{}
	err := m.db.QueryRowContext(ctx, query, uuid).Scan(&p.id, &p.productID, &p.UUID, &p.W, &p.H, &p.Path, &p.Typ, &p.Ori, &p.Up, &p.Pri, &p.Size, &p.Q, &p.GSURL, &p.Data, &p.Created, &p.Modified)
	if err != nil {
		return nil, err
	}
	return &p, nil
}

// DeleteProductImage deletes an image entry row from the product_image
// table by UUID.
func (m *PgModel) DeleteProductImage(ctx context.Context, uuid string) (int64, error) {
	query := `
		DELETE FROM product_image
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

// DeleteAllProductImages Images deletes all images from the product_image table
//  associated to the product with the given uuid.
func (m *PgModel) DeleteAllProductImages(ctx context.Context, productUUID string) error {
	tx, err := m.db.BeginTx(ctx, nil)
	if err != nil {
		return errors.Wrap(err, "db.BeginTx")
	}

	q1 := `SELECT id FROM product WHERE uuid = $1`
	var productID int
	err = tx.QueryRowContext(ctx, q1, productUUID).Scan(&productID)
	if err != nil {
		if err == sql.ErrNoRows {
			tx.Rollback()
			return ErrProductNotFound
		}
		tx.Rollback()
		return errors.Wrapf(err, "query row context failed for query=%q", q1)
	}

	q2 := `
		DELETE FROM product_image
		WHERE product_id = $1
	`
	_, err = m.db.ExecContext(ctx, q2, productID)
	if err != nil {
		tx.Rollback()
		return err
	}

	if err = tx.Commit(); err != nil {
		return errors.Wrap(err, "tx.Commit")
	}
	return nil
}
