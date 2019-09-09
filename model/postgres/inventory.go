package postgres

import (
	"context"
	"database/sql"
	"time"

	"github.com/pkg/errors"
)

// ErrInventoryNotFound is returned when any query
// for inventory returns no results.
var ErrInventoryNotFound = errors.New("postgres: inventory not found")

// InventoryRowUpdate holds the data for a single update used in batch update.
type InventoryRowUpdate struct {
	ProductUUID string
	Onhand      int
}

// A InventoryJoinRow represents a single row from the inventory table
// joined to the product table.
type InventoryJoinRow struct {
	id          int
	UUID        string
	productID   int
	ProductUUID string
	ProductPath string
	ProductSKU  string
	Onhand      int
	Created     time.Time
	Modified    time.Time
}

// GetInventoryByUUID returns a single InventoryJoinRow for a given inventory id.
func (m *PgModel) GetInventoryByUUID(ctx context.Context, inventoryUUID string) (*InventoryJoinRow, error) {
	q1 := `
		SELECT
		  v.id, v.uuid, v.product_id, p.uuid AS product_uuid,
		  p.path AS product_path, p.sku AS product_sku,
		  onhand, v.created, v.modified
		FROM inventory AS v
		INNER JOIN product AS p
		ON v.product_id = p.id
		WHERE v.uuid = $1
	`
	row := m.db.QueryRowContext(ctx, q1, inventoryUUID)
	var v InventoryJoinRow
	if err := row.Scan(&v.id, &v.UUID, &v.productID, &v.ProductUUID, &v.ProductPath, &v.ProductSKU, &v.Onhand, &v.Created, &v.Modified); err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrInventoryNotFound
		}
		return nil, errors.Wrap(err, "postgres: scan failed")
	}
	return &v, nil
}

// GetInventoryByProductUUID returns a single InventoryJoinRow for a given product id.
func (m *PgModel) GetInventoryByProductUUID(ctx context.Context, productUUID string) (*InventoryJoinRow, error) {
	q1 := `
		SELECT
		  v.id, v.uuid, v.product_id, p.uuid AS product_uuid,
		  p.path AS product_path, p.sku AS product_sku,
		  onhand, v.created, v.modified
		FROM inventory AS v
		INNER JOIN product AS p
		ON v.product_id = p.id
		WHERE p.uuid = $1
	`
	var v InventoryJoinRow
	err := m.db.QueryRowContext(ctx, q1, productUUID).Scan(&v.id, &v.UUID, &v.productID, &v.ProductUUID, &v.ProductPath, &v.ProductSKU, &v.Onhand, &v.Created, &v.Modified)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrProductNotFound
		}
		return nil, errors.Wrapf(err, "postgres: query row context failed for q1=%q", q1)
	}

	return &v, nil
}

// GetAllInventory returns a list of InventoryJoinRows for all inventory.
func (m *PgModel) GetAllInventory(ctx context.Context) ([]*InventoryJoinRow, error) {
	q1 := `
		SELECT
		  v.id, v.uuid, v.product_id, p.uuid AS product_uuid,
		  p.path AS product_path, p.sku AS product_sku,
		  onhand, v.created, v.modified
		FROM inventory AS v
		INNER JOIN product AS p
		ON v.product_id = p.id
	`
	rows, err := m.db.QueryContext(ctx, q1)
	if err != nil {
		return nil, errors.Wrapf(err, "postgres: m.db.QueryContext(ctx) failed")
	}
	defer rows.Close()

	list := make([]*InventoryJoinRow, 0, 128)
	for rows.Next() {
		var v InventoryJoinRow
		if err = rows.Scan(&v.id, &v.UUID, &v.productID, &v.ProductUUID, &v.ProductPath, &v.ProductSKU, &v.Onhand, &v.Created, &v.Modified); err != nil {
			if err == sql.ErrNoRows {
				return nil, ErrInventoryNotFound
			}
			return nil, errors.Wrap(err, "postgres: scan failed")
		}
		list = append(list, &v)
	}
	if err = rows.Err(); err != nil {
		return nil, errors.Wrap(err, "postgres: rows.Err()")
	}

	return list, nil
}

// UpdateInventoryByUUID updates the inventory with the given uuid
// returning the new inventory.
func (m *PgModel) UpdateInventoryByUUID(ctx context.Context, inventoryUUID string, onhand int) (*InventoryJoinRow, error) {
	tx, err := m.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, errors.Wrap(err, "postgres: db.BeginTx failed")
	}

	// 1. Check the inventory exists
	q1 := "SELECT id FROM inventory WHERE uuid = $1"
	var inventoryID int
	err = tx.QueryRowContext(ctx, q1, inventoryUUID).Scan(&inventoryID)
	if err != nil {
		if err == sql.ErrNoRows {
			tx.Rollback()
			return nil, ErrInventoryNotFound
		}
		tx.Rollback()
		return nil, errors.Wrapf(err, "postgres: query row context failed for q1=%q", q1)
	}

	// 2. Update the inventory row with the new onhold value.
	q2 := `
		UPDATE inventory SET onhand = $1, modified = NOW()
		WHERE id = $2
		RETURNING
		  id, uuid, product_id, onhand, created, modified
	`
	row := tx.QueryRowContext(ctx, q2, onhand, inventoryID)
	v := InventoryJoinRow{}
	if err := row.Scan(&v.id, &v.UUID, &v.productID, &v.Onhand, &v.Created, &v.Modified); err != nil {
		tx.Rollback()
		return nil, errors.Wrapf(err, "postgres: query row context q2=%q failed", q2)
	}

	q3 := "SELECT uuid, path, sku FROM product WHERE id = $1"
	row = tx.QueryRowContext(ctx, q3, v.productID)
	if err = row.Scan(&v.ProductUUID, &v.ProductPath, &v.ProductSKU); err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrProductNotFound
		}
		return nil, errors.Wrapf(err, "postgres: scan failed")
	}

	if err = tx.Commit(); err != nil {
		return nil, errors.Wrap(err, "postgres: tx.Commit failed")
	}

	return &v, nil
}

// BatchUpdateInventory updates multiple product inventory, either
// all completing or none.
func (m *PgModel) BatchUpdateInventory(ctx context.Context, inventoryList []*InventoryRowUpdate) ([]*InventoryJoinRow, error) {
	tx, err := m.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, errors.Wrap(err, "postgres: db.BeginTx failed")
	}

	// 1. Create a map of product uuid to product ids
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

	// Iterate the inventory update list passed in to this function.
	// Ensure that each inventory product exists in the map
	for _, inv := range inventoryList {
		if _, ok := productMap[inv.ProductUUID]; !ok {
			tx.Rollback()
			return nil, ErrProductNotFound
		}
	}

	q2 := `
		UPDATE inventory
		SET onhand = $1, modified = NOW()
		WHERE product_id = $2
		RETURNING id, uuid, product_id, onhand, created, modified
	`
	stmt2, err := tx.PrepareContext(ctx, q2)
	if err != nil {
		tx.Rollback()
		return nil, errors.Wrapf(err, "postgres: tx prepare for q2=%q", q2)
	}
	defer stmt2.Close()

	inventoryResults := make([]*InventoryJoinRow, 0, len(inventoryList))
	for _, i := range inventoryList {
		product := productMap[i.ProductUUID]

		inventory := InventoryJoinRow{}
		if err := stmt2.QueryRowContext(ctx, i.Onhand, product.id).Scan(&inventory.id, &inventory.UUID, &inventory.productID, &inventory.Onhand, &inventory.Created, &inventory.Modified); err != nil {
			if err == sql.ErrNoRows {
				tx.Rollback()
				return nil, ErrProductCategoryNotFound
			}
			tx.Rollback()
			return nil, errors.Wrapf(err, "postgres: stmt2.QueryRowContext(ctx, ...) failed q2%q", q2)
		}
		inventory.ProductUUID = product.uuid
		inventory.ProductPath = product.path
		inventory.ProductSKU = product.sku

		inventoryResults = append(inventoryResults, &inventory)
	}

	if err = tx.Commit(); err != nil {
		return nil, errors.Wrap(err, "postgres: tx.Commit failed")
	}

	return inventoryResults, nil
}
