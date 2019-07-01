package postgres

import (
	"context"
	"database/sql"
	"strconv"
	"time"

	"github.com/pkg/errors"
)

var (
	ErrCartItemAlreadyExists = errors.New("cart already exists")
	ErrCartItemNotFound      = errors.New("cart item not found")
)

// CartItem structure holds the details individual cart item.
type CartItem struct {
	id        int
	UUID      string
	SKU       string
	Qty       int
	UnitPrice float64
	Created   time.Time
	Modified  time.Time
}

// CreateCart creates a new shopping cart
func (m *PgModel) CreateCart(ctx context.Context) (*string, error) {
	var uuid string
	query := `SELECT UUID_GENERATE_V4() AS uuid`
	err := m.db.QueryRowContext(ctx, query).Scan(&uuid)
	if err != nil {
		return nil, errors.Wrapf(err, "db.QueryRow(ctx, %s)", query)
	}
	return &uuid, nil
}

// AddItemToCart adds a new item with sku, qty and unit price
func (m *PgModel) AddItemToCart(ctx context.Context, uuid, tierRef, sku string, qty int) (*CartItem, error) {
	item := CartItem{}

	// check if the item is already in the cart
	query := `SELECT EXISTS(SELECT 1 FROM carts WHERE uuid=$1 AND sku=$2) AS exists;`
	var exists bool
	m.db.QueryRowContext(ctx, query, uuid, sku).Scan(&exists)
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
		INSERT INTO carts (uuid, sku, qty, unit_price)
		VALUES ($1, $2, $3, $4)
		RETURNING id, uuid, sku, qty, unit_price, created, modified
	`
	row := m.db.QueryRowContext(ctx, query, uuid, sku, qty, unitPrice)
	if err := row.Scan(&item.id, &item.UUID, &item.SKU, &item.Qty,
		&item.UnitPrice, &item.Created, &item.Modified); err != nil {
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
		if err = rows.Scan(&c.id, &c.UUID, &c.SKU, &c.Qty, &c.UnitPrice, &c.Created, &c.Modified); err != nil {
			return nil, errors.Wrapf(err, "scan cart item %v", c)
		}
		cartItems = append(cartItems, &c)
	}
	if err = rows.Err(); err != nil {
		return nil, errors.Wrapf(err, "rows err")
	}
	return cartItems, nil
}

// UpdateItemByCartID updates the qty of a cart item of the given sku.
func (m *PgModel) UpdateItemByCartID(ctx context.Context, id, sku string, qty int) (*CartItem, error) {
	query := `
		UPDATE carts
		SET qty = $1, modified = NOW()
		WHERE uuid = $2 AND sku = $3
		RETURNING id, uuid, sku, qty, unit_price, created, modified
	`
	item := CartItem{}
	row := m.db.QueryRowContext(ctx, query, qty, id, sku)
	if err := row.Scan(&item.id, &item.UUID, &item.SKU, &item.Qty, &item.UnitPrice, &item.Created, &item.Modified); err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrCartItemNotFound
		}
		return nil, errors.Wrapf(err, "query row scan query=%q", query)
	}
	return &item, nil
}

// DeleteCartItem deletes a single cart item.
func (m *PgModel) DeleteCartItem(ctx context.Context, id, sku string) (count int64, err error) {
	query := `DELETE FROM carts WHERE uuid = $1 AND sku = $2`
	res, err := m.db.ExecContext(ctx, query, id, sku)
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
func (m *PgModel) EmptyCartItems(ctx context.Context, id string) (err error) {
	query := `DELETE FROM carts WHERE uuid = $1`
	_, err = m.db.ExecContext(ctx, query, id)
	if err != nil {
		return errors.Wrapf(err, "exec context query=%q", query)
	}
	return nil
}
