package postgres

import (
	"context"
	"database/sql"
	"strings"
	"time"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

// PriceEntry contains a tier reference and unit price pair
type PriceEntry struct {
	PriceListUUID string
	UnitPrice     int
}

// ProductUpdate contains the data required to update an existing product.
type ProductUpdate struct {
	Path string
	SKU  string
	Name string
}

// ProductRow maps to a product row.
type ProductRow struct {
	id       int
	UUID     string
	Path     string
	SKU      string
	Name     string
	Created  time.Time
	Modified time.Time
}

// ProductJoinRow represents a product row joined with a image row
type ProductJoinRow struct {
	id       int
	UUID     string
	Path     string
	SKU      string
	Name     string
	Images   []*ImageJoinRow
	Prices   []*PriceRow
	Created  time.Time
	Modified time.Time
}

// ErrProductNotFound is returned by GetProduct when the query
// for the product could not be found in the database.
var ErrProductNotFound = errors.New("postgres: product not found")

// ErrProductPathExists error
var ErrProductPathExists = errors.New("postgres: product path exists")

// ErrProductSKUExists error
var ErrProductSKUExists = errors.New("postgres: product sku exists")

// GetProduct returns a ProductRow by product id.
func (m *PgModel) GetProduct(ctx context.Context, productID string) (*ProductRow, error) {
	q1 := `
		SELECT id, uuid, sku, path, name, created, modified
		FROM product WHERE uuid = $1
	`
	p := ProductRow{}
	row := m.db.QueryRowContext(ctx, q1, productID)
	if err := row.Scan(&p.id, &p.UUID, &p.SKU, &p.Path, &p.Name, &p.Created, &p.Modified); err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrProductNotFound
		}
		return nil, errors.Wrapf(err, "postgres: query scan context productID=%q q1=%q failed", productID, q1)
	}
	return &p, nil
}

// GetProducts returns a list of all products in the product table.
func (m *PgModel) GetProducts(ctx context.Context) ([]*ProductRow, error) {
	query := "SELECT id, uuid, sku, path, name, created, modified FROM product"
	rows, err := m.db.QueryContext(ctx, query)
	if err != nil {
		return nil, errors.Wrapf(err, "postgres: m.db.QueryContext(ctx) query=%q failed", query)
	}
	defer rows.Close()

	products := make([]*ProductRow, 0, 256)
	for rows.Next() {
		var p ProductRow
		if err := rows.Scan(&p.id, &p.UUID, &p.SKU, &p.Path, &p.Name, &p.Created, &p.Modified); err != nil {
			return nil, errors.Wrap(err, "postgres: scan failed")
		}
		products = append(products, &p)
	}
	if err = rows.Err(); err != nil {
		return nil, errors.Wrap(err, "postgres: rows.Err() failed")
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

// CreateProduct updates the details of a product with the given product id.
func (m *PgModel) CreateProduct(ctx context.Context, userUUID string, path, sku, name string) (*ProductRow, error) {
	contextLogger := log.WithContext(ctx)
	contextLogger.Debugf("postgres: CreateProduct(ctx, userUUID=%s, path=%s, sku=%s, name=%q) called", userUUID, path, sku, name)

	tx, err := m.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, errors.Wrap(err, "postgres: db.BeginTx failed")
	}

	q1 := "SELECT EXISTS(SELECT 1 FROM product WHERE path = $1) AS exists"
	var exists bool
	err = tx.QueryRowContext(ctx, q1, path).Scan(&exists)
	if err != nil {
		return nil, errors.Wrapf(err, "postgres: tx.QueryRowContext(ctx, q1=%q, path=%q) failed", q1, path)
	}
	if exists {
		return nil, ErrProductPathExists
	}

	q2 := "SELECT EXISTS(SELECT 1 FROM product WHERE sku = $1) AS exists"
	err = tx.QueryRowContext(ctx, q2, sku).Scan(&exists)
	if err != nil {
		return nil, errors.Wrapf(err, "postgres: tx.QueryRowContext(ctx, q2=%q, path=%q) failed", q2, sku)
	}
	if exists {
		return nil, ErrProductSKUExists
	}

	// 3. Determine the price list the user is on.
	var priceListID int
	if userUUID != "" {
		q3 := "SELECT price_list_id FROM usr WHERE uuid = $1"
		err = tx.QueryRowContext(ctx, q3, userUUID).Scan(&priceListID)
		if err != nil {
			if err == sql.ErrNoRows {
				tx.Rollback()
				return nil, ErrUserNotFound
			}
			tx.Rollback()
			return nil, errors.Wrapf(err, "postgres: query row context failed for query=%q", q3)
		}
		contextLogger.Debugf("postgres: userUUID=%s is on priceListID=%d", userUUID, priceListID)
	} else {
		contextLogger.Debugf("postgres: userUUID not set. Trying to determine priceListID using the default price list code")
		q4 := "SELECT id FROM price_list WHERE code = 'default'"
		err = tx.QueryRowContext(ctx, q4).Scan(&priceListID)
		if err != nil {
			if err == sql.ErrNoRows {
				tx.Rollback()
				return nil, ErrDefaultPriceListNotFound
			}
			tx.Rollback()
			return nil, errors.Wrapf(err, "postgres: query row context failed for query=%q", q4)
		}
	}

	q4 := `
		INSERT INTO product
		  (path, sku, name, created, modified)
		VALUES
		  ($1, $2, $3, NOW(), NOW())
		RETURNING
		  id, uuid, path, sku, name, created, modified
	`
	p := ProductRow{}
	row := tx.QueryRowContext(ctx, q4, path, sku, name)
	if err := row.Scan(&p.id, &p.UUID, &p.Path, &p.SKU, &p.Name, &p.Created, &p.Modified); err != nil {
		tx.Rollback()
		return nil, errors.Wrapf(err, "postgres: query row context q4=%q failed", q4)
	}
	contextLogger.Debugf("postgres: q4 created new product with product.id=%d, product.UUID=%s", p.id, p.UUID)

	q5 := `
		INSERT INTO price (product_id, price_list_id, break, unit_price)
		VALUES ($1, $2, 1, 0)
	`
	_, err = tx.ExecContext(ctx, q5, p.id, priceListID)
	if err != nil {
		return nil, errors.Wrapf(err, "postgres: tx.ExecContext(ctx, q5=%q, %d)", q5, p.id)
	}
	contextLogger.Debugf("postgres: q5 created a single empty price for product.id=%d, p.UUID=%s", p.id, p.UUID)

	// 6. Set the inventory for this product to onhold = 0
	q6 := `
		INSERT INTO inventory (product_id, onhand, created, modified)
		VALUES ($1, 0, NOW(), NOW())
	`
	_, err = tx.ExecContext(ctx, q6, p.id)
	if err != nil {
		return nil, errors.Wrapf(err, "postgres: tx.ExecContext(ctx, q6=%q, %d)", q6, p.id)
	}

	// Delete all existing products. This is not the most efficient
	// way, but is easier that comparing the state of a list of new
	// images with the underlying database. Product updates don't
	// effect the user's experience.
	// q2 := "DELETE FROM image WHERE product_id = $1"
	// if _, err = tx.ExecContext(ctx, q2, p.id); err != nil {
	// 	tx.Rollback()
	// 	return nil, errors.Wrapf(err, "model: delete image query=%q", q2)
	// }

	// The inner sub select uses a pri of 10 if now rows exist for the given product
	// or pri + 10 for each subseqent row.
	// q3 := `
	// 	INSERT INTO image (
	// 		product_id,
	// 		w, h, path, typ,
	// 		ori, up,
	// 		pri, size, q,
	// 		gsurl, created, modified
	// 	) VALUES (
	// 		$1,
	// 		$2, $3, $4, $5,
	// 		$6, false,
	// 		(
	// 			SELECT
	// 				CASE WHEN COUNT(1) > 0
	// 				THEN MAX(pri)+10
	// 				ELSE 10
	// 			END
	// 			AS pri
	// 			FROM image
	// 			WHERE id = $7
	// 		), $8, $9,
	// 		$10, NOW(), NOW()
	// 	) RETURNING
	// 	  id, uuid, product_id, w,
	// 	  h, path, typ, ori, up, pri, size, q,
	// 	  gsurl, data, created, modified
	// `
	// stmt3, err := tx.PrepareContext(ctx, q3)
	// if err != nil {
	// 	tx.Rollback()
	// 	return nil, errors.Wrapf(err, "tx prepare for query=%q", q3)
	// }
	// defer stmt3.Close()

	// images := make([]*ImageJoinRow, 0)
	// for _, img := range pu.Images {
	// 	var pi ImageJoinRow
	// 	row := stmt3.QueryRowContext(ctx, p.id,
	// 		img.W, img.H, img.Path, img.Typ,
	// 		img.Ori,
	// 		p.id, img.Size, img.Q,
	// 		img.GSURL)
	// 	if err := row.Scan(&pi.id, &pi.UUID, &pi.productID, &pi.W,
	// 		&pi.H, &pi.Path, &pi.Typ, &pi.Ori, &pi.Up, &pi.Pri, &pi.Size, &pi.Q,
	// 		&pi.GSURL, &pi.Data, &p.Created, &p.Modified); err != nil {
	// 		tx.Rollback()
	// 		return nil, errors.Wrapf(err, "row.Scan failed")
	// 	}
	// 	pi.ProductUUID = p.UUID
	// 	images = append(images, &pi)
	// }

	// insert or update product pricing matrix (uses and 'upsert' with ON CONFLICT)
	// q4 := "SELECT id FROM price_list WHERE uuid = $1"
	// stmt4, err := tx.PrepareContext(ctx, q4)
	// if err != nil {
	// 	tx.Rollback()
	// 	return nil, errors.Wrapf(err, "tx prepare for query=%q", q4)
	// }
	// defer stmt4.Close()

	// q5 := `
	// 	INSERT INTO price
	// 	  (price_list_id, product_id, unit_price, created, modified)
	// 	VALUES
	// 	  ($1, $2, $3, NOW(), NOW())
	// 	ON CONFLICT (price_list_id, product_id)
	// 	DO UPDATE
	// 	  SET
	// 	    unit_price = $4, modified = NOW()
	// 	  WHERE
	// 	    excluded.product_id = $5 AND excluded.price_list_id = $6
	// 	RETURNING
	// 	  id, uuid, price_list_id, product_id, unit_price, created, modified
	// `
	// stmt5, err := tx.PrepareContext(ctx, q5)
	// if err != nil {
	// 	tx.Rollback()
	// 	return nil, errors.Wrapf(err, "tx prepare for query=%q", q5)
	// }
	// defer stmt5.Close()

	// prices := make([]*PriceRow, 0, 4)
	// for _, r := range pu.Prices {
	// 	var priceListID int
	// 	if err := stmt4.QueryRowContext(ctx, r.PriceListUUID).Scan(&priceListID); err != nil {
	// 		if err == sql.ErrNoRows {
	// 			tx.Rollback()
	// 			return nil, ErrPriceListNotFound
	// 		}
	// 		tx.Rollback()
	// 	}

	// 	row = stmt5.QueryRowContext(ctx, priceListID, p.id, r.UnitPrice,
	// 		r.UnitPrice, p.id, priceListID)
	// 	var pp PriceRow
	// 	if err := row.Scan(&pp.id, &pp.UUID, &pp.priceListID, &pp.productID, &pp.UnitPrice, &pp.Created, &pp.Modified); err != nil {
	// 		return nil, errors.Wrap(err, "scan failed")
	// 	}
	// 	prices = append(prices, &pp)
	// }
	// productFull := ProductJoinRow{
	// 	id:       p.id,
	// 	UUID:     p.UUID,
	// 	SKU:      p.SKU,
	// 	Path:     p.Path,
	// 	Name:     p.Name,
	// 	Images:   images,
	// 	Prices:   prices,
	// 	Created:  p.Created,
	// 	Modified: p.Modified,
	// }
	if err = tx.Commit(); err != nil {
		return nil, errors.Wrap(err, "postgres: tx.Commit failed")
	}
	return &p, nil
}

// UpdateProduct updates the details of a product with the given product id.
func (m *PgModel) UpdateProduct(ctx context.Context, productUUID string, pu *ProductUpdate) (*ProductRow, error) {
	tx, err := m.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, errors.Wrap(err, "postgres: db.BeginTx failed")
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

	// The clause
	//   AND id != $2
	// checks all other paths but not our own.
	q2 := `SELECT EXISTS(SELECT 1 FROM product WHERE path = $1 AND id != $2) AS exists`
	var exists bool
	err = tx.QueryRowContext(ctx, q2, pu.Path, productID).Scan(&exists)
	if err != nil {
		return nil, errors.Wrapf(err, "model: tx.QueryRowContext(ctx, q2=%q, pu.Path=%q)", q2, pu.Path)
	}
	if exists {
		return nil, ErrProductPathExists
	}

	// The clause
	//   AND id != $2
	// checks all other paths but not our own.
	q3 := `SELECT EXISTS(SELECT 1 FROM product WHERE sku = $1 AND id != $2) AS exists`
	err = tx.QueryRowContext(ctx, q3, pu.SKU, productID).Scan(&exists)
	if err != nil {
		return nil, errors.Wrapf(err, "model: tx.QueryRowContext(ctx, q3=%q, pu.SKU=%q)", q3, pu.SKU)
	}
	if exists {
		return nil, ErrProductSKUExists
	}

	q4 := `
		UPDATE product
		SET
		  path = $1, sku = $2, name = $3, modified = NOW()
		WHERE
		  id = $4
		RETURNING
		  id, uuid, path, sku, name, created, modified`
	row := tx.QueryRowContext(ctx, q4, pu.Path, pu.SKU, pu.Name, productID)

	p := ProductRow{}
	if err := row.Scan(&p.id, &p.UUID, &p.Path, &p.SKU, &p.Name, &p.Created, &p.Modified); err != nil {
		tx.Rollback()
		return nil, errors.Wrapf(err, "postgres: query row context q4=%q failed", q4)
	}

	// Delete all existing products. This is not the most efficient
	// way, but is easier that comparing the state of a list of new
	// images with the underlying database. Product updates don't
	// effect the user's experience.
	// q2 := "DELETE FROM image WHERE product_id = $1"
	// if _, err = tx.ExecContext(ctx, q2, p.id); err != nil {
	// 	tx.Rollback()
	// 	return nil, errors.Wrapf(err, "model: delete image query=%q", q2)
	// }

	// The inner sub select uses a pri of 10 if now rows exist for the given product
	// or pri + 10 for each subseqent row.
	// q3 := `
	// 	INSERT INTO image (
	// 		product_id,
	// 		w, h, path, typ,
	// 		ori, up,
	// 		pri, size, q,
	// 		gsurl, created, modified
	// 	) VALUES (
	// 		$1,
	// 		$2, $3, $4, $5,
	// 		$6, false,
	// 		(
	// 			SELECT
	// 				CASE WHEN COUNT(1) > 0
	// 				THEN MAX(pri)+10
	// 				ELSE 10
	// 			END
	// 			AS pri
	// 			FROM image
	// 			WHERE id = $7
	// 		), $8, $9,
	// 		$10, NOW(), NOW()
	// 	) RETURNING
	// 	  id, product_id, uuid, w,
	// 	  h, path, typ, ori, up, pri, size, q,
	// 	  gsurl, data, created, modified
	// `
	// stmt3, err := tx.PrepareContext(ctx, q3)
	// if err != nil {
	// 	tx.Rollback()
	// 	return nil, errors.Wrapf(err, "tx prepare for query=%q", q3)
	// }
	// defer stmt3.Close()

	// images := make([]*ImageJoinRow, 0)
	// for _, img := range pu.Images {
	// 	var pi ImageJoinRow
	// 	row := stmt3.QueryRowContext(ctx, p.id,
	// 		img.W, img.H, img.Path, img.Typ,
	// 		img.Ori,
	// 		p.id, img.Size, img.Q,
	// 		img.GSURL)
	// 	if err := row.Scan(&pi.id, &pi.productID, &pi.UUID, &pi.W,
	// 		&pi.H, &pi.Path, &pi.Typ, &pi.Ori, &pi.Up, &pi.Pri, &pi.Size, &pi.Q,
	// 		&pi.GSURL, &pi.Data, &p.Created, &p.Modified); err != nil {
	// 		tx.Rollback()
	// 		return nil, errors.Wrapf(err, "row.Scan failed")
	// 	}
	// 	images = append(images, &pi)
	// }

	// insert or update product pricing matrix (uses and 'upsert' with ON CONFLICT)
	// q4 := "SELECT id FROM price_list WHERE uuid = $1"
	// stmt4, err := tx.PrepareContext(ctx, q4)
	// if err != nil {
	// 	tx.Rollback()
	// 	return nil, errors.Wrapf(err, "tx prepare for query=%q", q4)
	// }
	// defer stmt4.Close()

	// q5 := `
	// 	INSERT INTO prices
	// 	  (price_list_id, product_id, unit_price, created, modified)
	// 	VALUES
	// 	  ($1, $2, $3, NOW(), NOW())
	// 	ON CONFLICT (price_list_id, product_id)
	// 	DO UPDATE
	// 	  SET
	// 	    unit_price = $4, modified = NOW()
	// 	  WHERE
	// 	    excluded.product_id = $5 AND excluded.price_list_id = $6
	// 	RETURNING
	// 	  id, uuid, price_list_id, product_id, unit_price, created, modified
	// `
	// stmt5, err := tx.PrepareContext(ctx, q5)
	// if err != nil {
	// 	tx.Rollback()
	// 	return nil, errors.Wrapf(err, "tx prepare for query=%q", q5)
	// }
	// defer stmt5.Close()

	// prices := make([]*PriceRow, 0, 4)
	// for _, r := range pu.Prices {
	// 	var priceListID int
	// 	if err := stmt4.QueryRowContext(ctx, r.PriceListUUID).Scan(&priceListID); err != nil {
	// 		if err == sql.ErrNoRows {
	// 			tx.Rollback()
	// 			return nil, ErrPriceListNotFound
	// 		}
	// 		tx.Rollback()
	// 	}

	// 	row = stmt5.QueryRowContext(ctx, priceListID, p.id, r.UnitPrice,
	// 		r.UnitPrice, p.id, priceListID)
	// 	var pp PriceRow
	// 	if err := row.Scan(&pp.id, &pp.UUID, &pp.priceListID, &pp.productID, &pp.UnitPrice, &pp.Created, &pp.Modified); err != nil {
	// 		return nil, errors.Wrap(err, "scan failed")
	// 	}
	// 	prices = append(prices, &pp)
	// }
	// productFull := ProductJoinRow{
	// 	id:       p.id,
	// 	UUID:     p.UUID,
	// 	SKU:      p.SKU,
	// 	Path:     p.Path,
	// 	Name:     p.Name,
	// 	Images:   images,
	// 	Prices:   prices,
	// 	Created:  p.Created,
	// 	Modified: p.Modified,
	// }
	if err = tx.Commit(); err != nil {
		return nil, errors.Wrap(err, "postgres: tx.Commit failed")
	}
	return &p, nil
}

// DeleteProduct delete the product with the given UUID.
func (m *PgModel) DeleteProduct(ctx context.Context, productUUID string) error {
	tx, err := m.db.BeginTx(ctx, nil)
	if err != nil {
		return errors.Wrap(err, "postgres: db.BeginTx failed")
	}

	// 1. Check if the product exists
	q1 := "SELECT id FROM product WHERE uuid = $1"
	var productID int
	err = tx.QueryRowContext(ctx, q1, productUUID).Scan(&productID)
	if err != nil {
		if err == sql.ErrNoRows {
			tx.Rollback()
			return ErrProductNotFound
		}
		tx.Rollback()
		return errors.Wrapf(err, "postgres: query row context failed for q1=%q", q1)
	}

	// 2. TODO: check if the product is part of a promo rule targeting an individual product.
	// return an error if it is in use.

	// 3. Remove the product from all cart items (one to many)
	q3 := "DELETE FROM cart_product WHERE product_id = $1"
	_, err = tx.ExecContext(ctx, q3, productID)
	if err != nil {
		tx.Rollback()
		return errors.Wrapf(err, "postgres: exec context q3=%q failed", q3)
	}

	// 4. Unlink the product from the categories (many to many)
	q4 := "DELETE FROM product_category WHERE product_id = $1"
	_, err = tx.ExecContext(ctx, q4, productID)
	if err != nil {
		tx.Rollback()
		return errors.Wrapf(err, "postgres: exec context q4=%q failed", q4)
	}

	// 5. Remove all images associated with this product (one to many)
	q5 := "DELETE FROM image WHERE product_id = $1"
	_, err = tx.ExecContext(ctx, q5, productID)
	if err != nil {
		tx.Rollback()
		return errors.Wrapf(err, "postgres: exec context q5=%q failed", q5)
	}

	// 6. Remove the inventory for this product (one to one)
	q6 := "DELETE FROM inventory WHERE product_id = $1"
	_, err = tx.ExecContext(ctx, q6, productID)
	if err != nil {
		tx.Rollback()
		return errors.Wrapf(err, "postgres: exec context q6=%q failed", q6)
	}

	// 7. Delete all prices referencing this product (many to many)
	q7 := "DELETE FROM price WHERE product_id = $1"
	_, err = tx.ExecContext(ctx, q7, productID)
	if err != nil {
		tx.Rollback()
		return errors.Wrapf(err, "postgres: exec context q7=%q failed", q7)
	}

	// 8. Remove the product from any product sets to which it belongs (many to many)
	q8 := "DELETE FROM product_set_item WHERE product_id = $1"
	_, err = tx.ExecContext(ctx, q8, productID)
	if err != nil {
		tx.Rollback()
		return errors.Wrapf(err, "postgres: exec context q8=%q failed", q8)
	}

	// TODO: 9. Delete the empty product set.

	// 10. Delete the product
	q10 := "DELETE FROM product WHERE id = $1"
	_, err = tx.ExecContext(ctx, q10, productID)
	if err != nil {
		tx.Rollback()
		return errors.Wrapf(err, "exec context q10=%q", q10)
	}

	if err = tx.Commit(); err != nil {
		return errors.Wrap(err, "postgres: tx.Commit failed")
	}
	return nil
}
