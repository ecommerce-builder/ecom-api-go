package postgres

import (
	"context"
	"database/sql"
	"strconv"
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

// CartRow represents a row from the the carts table.
type CartRow struct {
	id       int
	UUID     string
	Locked   bool
	Created  time.Time
	Modified time.Time
}

// CartItemRow represents a row from the cart_items table.
type CartItemRow struct {
	id        int
	UUID      string
	cartID    int
	SKU       string
	Qty       int
	UnitPrice int
	Created   time.Time
	Modified  time.Time
}

// CartProductItem holds details of the an invidual cart item joined
// with product info.
type CartProductItem struct {
	id        int
	UUID      string
	SKU       string
	Name      string
	Qty       int
	UnitPrice int
	Created   time.Time
	Modified  time.Time
}

// CreateCart creates a new shopping cart
func (m *PgModel) CreateCart(ctx context.Context) (*CartRow, error) {
	var c CartRow
	query := `
		INSERT INTO carts (uuid, locked, created, modified)
		VALUES (UUID_GENERATE_V4(), 'f', NOW(), NOW())
		RETURNING id, uuid, locked, created, modified
		`
	row := m.db.QueryRowContext(ctx, query)
	if err := row.Scan(&c.id, &c.UUID, &c.Locked, &c.Created, &c.Modified); err != nil {
		return nil, errors.Wrapf(err, "query scan failed query=%q", query)
	}
	return &c, nil
}

// IsCartExists returns true if the cart with the given UUID exists.
func (m *PgModel) IsCartExists(ctx context.Context, cartUUID string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM carts WHERE uuid=$1) AS exists`
	var exists bool
	if err := m.db.QueryRowContext(ctx, query, cartUUID).Scan(&exists); err != nil {
		return false, errors.Wrapf(err, "query=%q scan failed", query)
	}
	if exists == true {
		return true, nil
	}
	return false, nil
}

// AddItemToCart adds a new item with sku, qty and unit price
func (m *PgModel) AddItemToCart(ctx context.Context, cartUUID, tierRef, sku string, qty int) (*CartProductItem, error) {
	item := CartProductItem{}

	// check if the item is already in the cart
	query := `
		SELECT EXISTS(
		  SELECT 1 FROM cart_items WHERE cart_id=(
		    SELECT id FROM carts WHERE uuid = $1
		  ) AND sku=$2
		) AS exists`
	var exists bool
	m.db.QueryRowContext(ctx, query, cartUUID, sku).Scan(&exists)
	if exists == true {
		return nil, ErrCartItemAlreadyExists
	}

	var unitPriceStr []byte
	query = `SELECT unit_price FROM product_pricing WHERE tier_ref = $1 AND sku = $2`
	err := m.db.QueryRowContext(ctx, query, tierRef, sku).Scan(&unitPriceStr)
	if err != nil {
		return &item, errors.Wrapf(err, "query scan failed query=%q", query)
	}

	unitPrice, _ := strconv.ParseFloat(string(unitPriceStr), 64)
	query = `
		INSERT INTO cart_items (cart_id, sku, qty, unit_price)
		VALUES ((SELECT id FROM carts WHERE uuid = $1), $2, $3, $4)
		RETURNING
		  id, uuid, sku, (SELECT name FROM products WHERE sku = $5),
		  qty, unit_price, created, modified
	`
	row := m.db.QueryRowContext(ctx, query, cartUUID, sku, qty, unitPrice, sku)
	if err := row.Scan(&item.id, &item.UUID, &item.SKU, &item.Name, &item.Qty,
		&item.UnitPrice, &item.Created, &item.Modified); err != nil {
		return nil, errors.Wrapf(err, "query scan failed query=%q", query)
	}
	return &item, nil
}

// HasCartItems returns true if any cart items has previously been added.
func (m *PgModel) HasCartItems(ctx context.Context, cartUUID string) (bool, error) {
	query := "SELECT COUNT(*) AS count FROM cart_items WHERE cart_id = (SELECT id FROM carts WHERE uuid = $1)"
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
func (m *PgModel) GetCartItems(ctx context.Context, cartUUID string) ([]*CartProductItem, error) {
	exists, err := m.IsCartExists(ctx, cartUUID)
	if err != nil {
		return nil, errors.Wrapf(err, "m.IsCartExists(ctx, cartUUID=%q) failed", cartUUID)
	}
	if !exists {
		return nil, ErrCartNotFound
	}

	query := `
		SELECT
		  C.id, C.uuid, C.sku, name, qty, unit_price, C.created, C.modified
		FROM cart_items AS C INNER JOIN products AS P
		  ON C.sku = P.sku
		WHERE
		  C.cart_id = (SELECT id FROM carts WHERE uuid = $1)
		ORDER BY created ASC
	`
	rows, err := m.db.QueryContext(ctx, query, cartUUID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	cartItems := make([]*CartProductItem, 0, 20)
	for rows.Next() {
		c := CartProductItem{}
		if err = rows.Scan(&c.id, &c.UUID, &c.SKU, &c.Name, &c.Qty, &c.UnitPrice, &c.Created, &c.Modified); err != nil {
			return nil, errors.Wrapf(err, "scan cart item %v", c)
		}
		cartItems = append(cartItems, &c)
	}
	if err = rows.Err(); err != nil {
		return nil, errors.Wrapf(err, "rows err")
	}
	return cartItems, nil
}

// UpdateItemByCartUUID updates the qty of a cart item of the given sku.
func (m *PgModel) UpdateItemByCartUUID(ctx context.Context, cartUUID, sku string, qty int) (*CartProductItem, error) {
	query := `
		UPDATE cart_items
		SET
		  qty = $1, modified = NOW()
		WHERE
		  cart_id = (SELECT id FROM carts WHERE uuid = $2) AND sku = $3
		RETURNING
		  id, uuid, sku, (SELECT name FROM products WHERE sku = $4),
		  qty, unit_price, created, modified
	`
	i := CartProductItem{}
	row := m.db.QueryRowContext(ctx, query, qty, cartUUID, sku, sku)
	if err := row.Scan(&i.id, &i.UUID, &i.SKU, &i.Name, &i.Qty, &i.UnitPrice, &i.Created, &i.Modified); err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrCartItemNotFound
		}
		return nil, errors.Wrapf(err, "query row scan query=%q", query)
	}
	return &i, nil
}

// DeleteCartItem deletes a single cart item.
func (m *PgModel) DeleteCartItem(ctx context.Context, cartUUID, sku string) (count int64, err error) {
	exists, err := m.IsCartExists(ctx, cartUUID)
	if err != nil {
		return 0, errors.Wrapf(err, "m.IsCartExists(ctx, cartUUID=%q) failed", cartUUID)
	}
	if !exists {
		return 0, ErrCartNotFound
	}
	query := `DELETE FROM cart_items WHERE cart_id = (SELECT id FROM carts WHERE uuid = $1) AND sku = $2`
	res, err := m.db.ExecContext(ctx, query, cartUUID, sku)
	if err != nil {
		return -1, errors.Wrapf(err, "exec context query=%q", query)
	}
	count, err = res.RowsAffected()
	if err != nil {
		return -1, errors.Wrap(err, "rows affected")
	}
	return count, nil
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

	query := `DELETE FROM cart_items WHERE cart_id = (SELECT id FROM carts WHERE uuid = $1)`
	_, err = m.db.ExecContext(ctx, query, cartUUID)
	if err != nil {
		return errors.Wrapf(err, "exec context query=%q", query)
	}
	return nil
}
