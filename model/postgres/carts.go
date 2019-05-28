package postgres

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/pkg/errors"
)

// CartItem structure holds the details individual cart item.
type CartItem struct {
	ID        int
	CartUUID  string
	Sku       string
	Qty       int
	UnitPrice float64
	Created   time.Time
	Modified  time.Time
}

// CreateCart creates a new shopping cart
func (m *PgModel) CreateCart(ctx context.Context) (*string, error) {
	var cartUUID string
	query := `SELECT UUID_GENERATE_V4() AS uuid`
	err := m.db.QueryRowContext(ctx, query).Scan(&cartUUID)
	if err != nil {
		return nil, errors.Wrapf(err, "db.QueryRow(ctx, %s)", query)
	}
	return &cartUUID, nil
}

// AddItemToCart adds a new item with sku, qty and unit price
func (m *PgModel) AddItemToCart(ctx context.Context, cartUUID, tierRef, sku string, qty int) (*CartItem, error) {
	item := CartItem{}

	// check if the item is already in the cart
	query := `SELECT EXISTS(SELECT 1 FROM carts WHERE uuid=$1 AND sku=$2) AS exists;`
	var exists bool
	m.db.QueryRowContext(ctx, query, cartUUID, sku).Scan(&exists)
	if exists == true {
		return nil, fmt.Errorf("cart item %s already exists", sku)
	}

	var unitPriceStr []byte
	query = `SELECT unit_price FROM product_pricing WHERE tier_ref = $1 AND sku = $2`
	err := m.db.QueryRowContext(ctx, query, tierRef, sku).Scan(&unitPriceStr)
	if err != nil {
		return &item, errors.Wrapf(err, "query scan failed query=%q", query)
	}

	unitPrice, _ := strconv.ParseFloat(string(unitPriceStr), 64)
	query = `
		INSERT INTO carts (uuid, sku, qty, unit_price)
		VALUES ($1, $2, $3, $4)
		RETURNING id, uuid, sku, qty, unit_price, created, modified
	`

	err = m.db.QueryRowContext(ctx, query, cartUUID, sku, qty, unitPrice).Scan(&item.ID, &item.CartUUID, &item.Sku, &item.Qty, &item.UnitPrice, &item.Created, &item.Modified)
	if err != nil {
		return nil, errors.Wrapf(err, "query scan failed query=%q", query)
	}
	return &item, nil
}

// GetCartItems gets all items in the cart
func (m *PgModel) GetCartItems(ctx context.Context, cartUUID string) ([]*CartItem, error) {
	query := `
		SELECT
			id, uuid, sku, qty, unit_price, created, modified
		FROM carts
		WHERE uuid = $1
	`
	rows, err := m.db.QueryContext(ctx, query, cartUUID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	cartItems := make([]*CartItem, 0, 20)
	for rows.Next() {
		c := CartItem{}
		err = rows.Scan(&c.ID, &c.CartUUID, &c.Sku, &c.Qty, &c.UnitPrice, &c.Created, &c.Modified)
		if err != nil {
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
func (m *PgModel) UpdateItemByCartUUID(ctx context.Context, cartUUID, sku string, qty int) (*CartItem, error) {
	query := `
		UPDATE carts
		SET qty = $1, modified = NOW()
		WHERE uuid = $2 AND sku = $3
		RETURNING id, uuid, sku, qty, unit_price, created, modified
	`

	item := CartItem{}
	err := m.db.QueryRowContext(ctx, query, qty, cartUUID, sku).Scan(&item.ID, &item.CartUUID, &item.Sku, &item.Qty, &item.UnitPrice, &item.Created, &item.Modified)
	if err != nil {
		return nil, errors.Wrapf(err, "query row scan query=%q", query)
	}
	return &item, nil
}

// DeleteCartItem deletes a single cart item.
func (m *PgModel) DeleteCartItem(ctx context.Context, cartUUID, sku string) (count int64, err error) {
	query := `DELETE FROM carts WHERE uuid = $1 AND sku = $2`
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
	query := `DELETE FROM carts WHERE uuid = $1`
	_, err = m.db.ExecContext(ctx, query, cartUUID)
	if err != nil {
		return errors.Wrapf(err, "exec context query=%q", query)
	}
	return nil
}
