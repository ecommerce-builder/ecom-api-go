package postgres

import (
	"context"
	"database/sql"
	"time"

	"github.com/pkg/errors"
)

var (
	// ErrCartNotFound is returned when attempting an operation on non existing cart.
	ErrCartNotFound = errors.New("cart not found")

	// ErrCartItemAlreadyExists occurs when trying to add an item
	// that already exists.
	ErrCartItemAlreadyExists = errors.New("cart already exists")

	// ErrCartItemNotFound occurs if the cart UUID and product SKU don't
	// match an existing cart item.
	ErrCartItemNotFound = errors.New("cart item not found")

	// ErrCartContainsNoItems occurs when attempting to delete all items.
	ErrCartContainsNoItems = errors.New("cart contains no items")
)

// CartRow represents a row from the the cart table.
type CartRow struct {
	id       int
	UUID     string
	Locked   bool
	Created  time.Time
	Modified time.Time
}

// CartItemRow represents a row from the cart_item table.
type CartItemRow struct {
	id        int
	UUID      string
	cartID    int
	productID int
	Qty       int
	UnitPrice int
	Created   time.Time
	Modified  time.Time
}

// CartItemJoinRow holds details of the an invidual cart item joined
// with product info.
type CartItemJoinRow struct {
	id          int
	UUID        string
	CartUUID    string
	productID   string
	ProductUUID string
	SKU         string
	Name        string
	Qty         int
	UnitPrice   int
	Created     time.Time
	Modified    time.Time
}

// CreateCart creates a new shopping cart
func (m *PgModel) CreateCart(ctx context.Context) (*CartRow, error) {
	var c CartRow
	query := `
		INSERT INTO cart
		  (uuid, locked, created, modified)
		VALUES
		  (UUID_GENERATE_V4(), 'f', NOW(), NOW())
		RETURNING
		  id, uuid, locked, created, modified
		`
	row := m.db.QueryRowContext(ctx, query)
	if err := row.Scan(&c.id, &c.UUID, &c.Locked, &c.Created, &c.Modified); err != nil {
		return nil, errors.Wrapf(err, "query scan failed query=%q", query)
	}
	return &c, nil
}

// IsCartExists returns true if the cart with the given UUID exists.
func (m *PgModel) IsCartExists(ctx context.Context, cartUUID string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM cart WHERE uuid=$1) AS exists`
	var exists bool
	if err := m.db.QueryRowContext(ctx, query, cartUUID).Scan(&exists); err != nil {
		return false, errors.Wrapf(err, "query=%q scan failed", query)
	}
	if exists == true {
		return true, nil
	}
	return false, nil
}

// AddItemToCart adds a new item with productUUID and qty.
func (m *PgModel) AddItemToCart(ctx context.Context, cartUUID, customerUUID, productUUID string, qty int) (*CartItemJoinRow, error) {
	tx, err := m.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, errors.Wrap(err, "db.BeginTx")
	}

	q1 := "SELECT id FROM cart WHERE uuid = $1"
	var cartID int
	err = tx.QueryRowContext(ctx, q1, cartUUID).Scan(&cartID)
	if err != nil {
		if err == sql.ErrNoRows {
			tx.Rollback()
			return nil, ErrCartNotFound
		}
		tx.Rollback()
		return nil, errors.Wrapf(err, "query row context failed for query=%q", q1)
	}

	q2 := "SELECT id FROM customer WHERE uuid = $1"
	var customerID int
	err = tx.QueryRowContext(ctx, q2, customerUUID).Scan(&customerID)
	if err != nil {
		if err == sql.ErrNoRows {
			tx.Rollback()
			return nil, ErrCustomerNotFound
		}
		tx.Rollback()
		return nil, errors.Wrapf(err, "query row context failed for query=%q", q2)
	}

	q3 := "SELECT id FROM product WHERE uuid = $1"
	var productID int
	err = tx.QueryRowContext(ctx, q3, productUUID).Scan(&productID)
	if err != nil {
		if err == sql.ErrNoRows {
			tx.Rollback()
			return nil, ErrProductNotFound
		}
		tx.Rollback()
		return nil, errors.Wrapf(err, "query row context failed for query=%q", q3)
	}

	var priceListID int
	if customerUUID != "" {
		q4 := "SELECT price_list_id FROM customer WHERE uuid = $1"
		err = tx.QueryRowContext(ctx, q4, customerUUID).Scan(&priceListID)
		if err != nil {
			if err == sql.ErrNoRows {
				tx.Rollback()
				return nil, ErrCustomerNotFound
			}
			tx.Rollback()
			return nil, errors.Wrapf(err, "query row context failed for query=%q", q4)
		}
	} else {
		q4 := "SELECT id FROM price_list WHERE code = 'default'"
		err = tx.QueryRowContext(ctx, q4).Scan(&priceListID)
		if err != nil {
			if err == sql.ErrNoRows {
				tx.Rollback()
				return nil, ErrDefaultPriceListMissing
			}
			tx.Rollback()
			return nil, errors.Wrapf(err, "query row context failed for query=%q", q4)
		}
	}

	// check if the item is already in the cart
	q5 := `
		SELECT EXISTS(
		  SELECT 1 FROM cart_item WHERE cart_id = $1 AND product_id = $2
		) AS exists`
	var exists bool
	tx.QueryRowContext(ctx, q5, cartID, productID).Scan(&exists)
	if exists == true {
		tx.Rollback()
		return nil, ErrCartItemAlreadyExists
	}

	q6 := `
		INSERT INTO cart_item
		  (cart_id, product_id, qty)
		VALUES
		  ($1, $2, $3)
		RETURNING
		  id
	`
	var cartItemID int
	row := tx.QueryRowContext(ctx, q6, cartID, productID, qty)
	if err := row.Scan(&cartItemID); err != nil {
		tx.Rollback()
		return nil, errors.Wrapf(err, "query scan failed query=%q", q6)
	}

	q7 := `
		SELECT
		  c.id, c.uuid, c.product_id, p.uuid, sku, name, qty, unit_price, c.created, c.modified
		FROM cart_item AS c
		INNER JOIN product AS p
		  ON p.id = c.product_id
		LEFT OUTER JOIN price AS r
		  ON r.product_id = p.id
		WHERE
		  c.id = $1 AND r.price_list_id = $2
		ORDER BY created ASC
	`
	item := CartItemJoinRow{}
	row = tx.QueryRowContext(ctx, q7, cartItemID, priceListID)
	if err := row.Scan(&item.id, &item.UUID, &item.productID, &item.ProductUUID, &item.SKU, &item.Name,
		&item.Qty, &item.UnitPrice, &item.Created, &item.Modified); err != nil {
		tx.Rollback()
		return nil, errors.Wrapf(err, "query scan failed query=%q", q7)
	}
	item.CartUUID = cartUUID

	if err = tx.Commit(); err != nil {
		return nil, errors.Wrap(err, "tx.Commit")
	}
	return &item, nil
}

// HasCartItems returns true if any cart item has previously been added.
func (m *PgModel) HasCartItems(ctx context.Context, cartUUID string) (bool, error) {
	query := "SELECT COUNT(*) AS count FROM cart_item WHERE cart_id = (SELECT id FROM cart WHERE uuid = $1)"
	var count int
	err := m.db.QueryRowContext(ctx, query, cartUUID).Scan(&count)
	if err != nil {
		return false, errors.Wrapf(err, "query row context scan query=%q", query)
	}
	if count == 0 {
		return false, nil
	}
	return true, nil
}

// GetCartItems gets all items in the cart
func (m *PgModel) GetCartItems(ctx context.Context, cartUUID, customerUUID string) ([]*CartItemJoinRow, error) {
	tx, err := m.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, errors.Wrap(err, "db.BeginTx")
	}

	q1 := "SELECT id FROM cart WHERE uuid = $1"
	var cartID int
	err = tx.QueryRowContext(ctx, q1, cartUUID).Scan(&cartID)
	if err != nil {
		if err == sql.ErrNoRows {
			tx.Rollback()
			return nil, ErrCartNotFound
		}
		tx.Rollback()
		return nil, errors.Wrapf(err, "query row context failed for query=%q", q1)
	}

	q2 := "SELECT id FROM customer WHERE uuid = $1"
	var customerID int
	err = tx.QueryRowContext(ctx, q2, customerUUID).Scan(&customerID)
	if err != nil {
		if err == sql.ErrNoRows {
			tx.Rollback()
			return nil, ErrCustomerNotFound
		}
		tx.Rollback()
		return nil, errors.Wrapf(err, "query row context failed for query=%q", q2)
	}

	var priceListID int
	if customerUUID != "" {
		q3 := "SELECT price_list_id FROM customer WHERE uuid = $1"
		err = tx.QueryRowContext(ctx, q3, customerUUID).Scan(&priceListID)
		if err != nil {
			if err == sql.ErrNoRows {
				tx.Rollback()
				return nil, ErrCustomerNotFound
			}
			tx.Rollback()
			return nil, errors.Wrapf(err, "query row context failed for query=%q", q3)
		}
	} else {
		q4 := "SELECT id FROM price_list WHERE code = 'default'"
		err = tx.QueryRowContext(ctx, q4).Scan(&priceListID)
		if err != nil {
			if err == sql.ErrNoRows {
				tx.Rollback()
				return nil, ErrDefaultPriceListMissing
			}
			tx.Rollback()
			return nil, errors.Wrapf(err, "query row context failed for query=%q", q4)
		}
	}
	q5 := `
		SELECT
		  c.id, c.uuid, c.product_id, p.uuid as product_uuid, sku, name, c.qty, unit_price, c.created, c.modified
		FROM cart_item AS c
		INNER JOIN product AS p
		  ON p.id = c.product_id
		LEFT OUTER JOIN price AS r
		  ON r.product_id = p.id
		WHERE
		  c.cart_id = $1 AND r.price_list_id = $2
		ORDER BY created ASC
	`
	rows, err := tx.QueryContext(ctx, q5, cartID, priceListID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	cartItems := make([]*CartItemJoinRow, 0, 20)
	for rows.Next() {
		c := CartItemJoinRow{}
		if err = rows.Scan(&c.id, &c.UUID, &c.productID, &c.ProductUUID, &c.SKU, &c.Name, &c.Qty, &c.UnitPrice, &c.Created, &c.Modified); err != nil {
			return nil, errors.Wrapf(err, "scan cart item %v", c)
		}
		c.CartUUID = cartUUID
		cartItems = append(cartItems, &c)
	}
	if err = rows.Err(); err != nil {
		return nil, errors.Wrapf(err, "rows err")
	}

	if err = tx.Commit(); err != nil {
		return nil, errors.Wrap(err, "tx.Commit")
	}
	return cartItems, nil
}

// UpdateItemByCartUUID updates the qty of a cart item of the given product id.
func (m *PgModel) UpdateItemByCartUUID(ctx context.Context, cartUUID, customerUUID, productUUID string, qty int) (*CartItemJoinRow, error) {
	tx, err := m.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, errors.Wrap(err, "db.BeginTx")
	}

	q1 := "SELECT id FROM cart WHERE uuid = $1"
	var cartID int
	err = tx.QueryRowContext(ctx, q1, cartUUID).Scan(&cartID)
	if err != nil {
		if err == sql.ErrNoRows {
			tx.Rollback()
			return nil, ErrCartNotFound
		}
		tx.Rollback()
		return nil, errors.Wrapf(err, "query row context failed for q1=%q", q1)
	}

	q2 := "SELECT id FROM product WHERE uuid = $1"
	var productID int
	if err = tx.QueryRowContext(ctx, q2, productUUID).Scan(&productID); err != nil {
		if err == sql.ErrNoRows {
			tx.Rollback()
			return nil, ErrProductNotFound
		}
		tx.Rollback()
		return nil, errors.Wrapf(err, "query row context failed for q2=%q", q2)
	}

	var priceListID int
	if customerUUID != "" {
		q3 := "SELECT price_list_id FROM customer WHERE uuid = $1"
		err = tx.QueryRowContext(ctx, q3, customerUUID).Scan(&priceListID)
		if err != nil {
			if err == sql.ErrNoRows {
				tx.Rollback()
				return nil, ErrCustomerNotFound
			}
			tx.Rollback()
			return nil, errors.Wrapf(err, "query row context failed for query=%q", q3)
		}
	} else {
		q3 := "SELECT id FROM price_list WHERE code = 'default'"
		err = tx.QueryRowContext(ctx, q3).Scan(&priceListID)
		if err != nil {
			if err == sql.ErrNoRows {
				tx.Rollback()
				return nil, ErrDefaultPriceListMissing
			}
			tx.Rollback()
			return nil, errors.Wrapf(err, "query row context failed for query=%q", q3)
		}
	}

	q4 := `
		UPDATE cart_item
		SET
		  qty = $1, modified = NOW()
		WHERE
		  cart_id = $2 AND product_id = $3
		RETURNING id
	`
	var cartItemID int
	row := tx.QueryRowContext(ctx, q4, qty, cartID, productID)
	if err := row.Scan(&cartItemID); err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrCartItemNotFound
		}
		tx.Rollback()
		return nil, errors.Wrapf(err, "query row context failed for q4=%q", q4)
	}

	q5 := `
		SELECT
		  c.id, c.uuid, c.product_id, p.uuid, sku, name, qty, unit_price, c.created, c.modified
		FROM cart_item AS c
		INNER JOIN product AS p
		  ON p.id = c.product_id
		LEFT OUTER JOIN price AS r
		  ON r.product_id = p.id
		WHERE
		  c.id = $1 AND r.price_list_id = $2
		ORDER BY created ASC
	`
	item := CartItemJoinRow{}
	row = tx.QueryRowContext(ctx, q5, cartItemID, priceListID)
	if err := row.Scan(&item.id, &item.UUID, &item.productID, &item.ProductUUID, &item.SKU, &item.Name,
		&item.Qty, &item.UnitPrice, &item.Created, &item.Modified); err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrCartItemNotFound
		}
		return nil, errors.Wrapf(err, "query row scan failed for q5=%q", q5)
	}

	if err = tx.Commit(); err != nil {
		return nil, errors.Wrap(err, "tx.Commit")
	}
	return &item, nil
}

// DeleteCartItem deletes a single cart item. Return `ErrCartNotFound` if
// the cart cannot be found, or `ErrProductNotFound` if the product is not
// in the product table. If the product is not in the cart it returns
/// `ErrCartItemNotFound`.
func (m *PgModel) DeleteCartItem(ctx context.Context, cartUUID, productUUID string) error {
	tx, err := m.db.BeginTx(ctx, nil)
	if err != nil {
		return errors.Wrap(err, "db.BeginTx")
	}

	q1 := "SELECT id FROM cart WHERE uuid = $1"
	var cartID int
	if err = tx.QueryRowContext(ctx, q1, cartUUID).Scan(&cartID); err != nil {
		if err == sql.ErrNoRows {
			tx.Rollback()
			return ErrCartNotFound
		}
		tx.Rollback()
		return errors.Wrapf(err, "query row context failed for q1=%q", q1)
	}

	q2 := "SELECT id FROM product WHERE uuid = $1"
	var productID int
	if err = tx.QueryRowContext(ctx, q2, productUUID).Scan(&productID); err != nil {
		if err == sql.ErrNoRows {
			tx.Rollback()
			return ErrProductNotFound
		}
		tx.Rollback()
		return errors.Wrapf(err, "query row context failed for q2=%q", q2)
	}

	q3 := "SELECT id FROM cart_item WHERE cart_id = $1 AND product_id = $2"
	var cartItemID int
	if err = tx.QueryRowContext(ctx, q3, cartID, productID).Scan(&cartItemID); err != nil {
		if err == sql.ErrNoRows {
			tx.Rollback()
			return ErrCartItemNotFound
		}
		tx.Rollback()
		return errors.Wrapf(err, "query row context failed for q3=%q", q3)
	}

	q4 := "DELETE FROM cart_item WHERE id = $1"
	_, err = tx.ExecContext(ctx, q4, cartItemID)
	if err != nil {
		return errors.Wrapf(err, "exec context q4=%q", q4)
	}

	if err = tx.Commit(); err != nil {
		return errors.Wrap(err, "tx.Commit")
	}

	return nil
}

// EmptyCartItems empty the cart of all items. Does not affect coupons.
func (m *PgModel) EmptyCartItems(ctx context.Context, cartUUID string) (err error) {
	exists, err := m.IsCartExists(ctx, cartUUID)
	if err != nil {
		return errors.Wrapf(err, "m.IsCartExists(ctx, cartUUID=%q) failed", cartUUID)
	}
	if !exists {
		return ErrCartNotFound
	}

	hasItems, err := m.HasCartItems(ctx, cartUUID)
	if err != nil {
		return errors.Wrapf(err, "m.HasCartItems(ctx, cartUUID=%q) failed: %v", cartUUID, err)
	}

	if !hasItems {
		return ErrCartContainsNoItems
	}

	query := `DELETE FROM cart_item WHERE cart_id = (SELECT id FROM cart WHERE uuid = $1)`
	_, err = m.db.ExecContext(ctx, query, cartUUID)
	if err != nil {
		return errors.Wrapf(err, "exec context query=%q", query)
	}
	return nil
}
