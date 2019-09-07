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

	// ErrCartProductExists occurs when trying to add an product
	// that already exists.
	ErrCartProductExists = errors.New("postgres: cart already exists")

	// ErrCartProductNotFound occurs if the cart product cannot be found.
	ErrCartProductNotFound = errors.New("postgres: cart product not found")

	// ErrCartContainsNoProducts occurs when attempting to delete all products from a cart.
	ErrCartContainsNoProducts = errors.New("postgres: cart contains no products")

	// ErrProductHasNoPrices error
	ErrProductHasNoPrices = errors.New("postgres: product has no prices")
)

// CartRow represents a row from the the cart table.
type CartRow struct {
	id       int
	UUID     string
	Locked   bool
	Created  time.Time
	Modified time.Time
}

// CartProductRow represents a row from the cart_product table.
type CartProductRow struct {
	id        int
	UUID      string
	cartID    int
	productID int
	Qty       int
	UnitPrice int
	Created   time.Time
	Modified  time.Time
}

// CartProductJoinRow holds details of the an invidual cart product joined
// with product info.
type CartProductJoinRow struct {
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
		return nil, errors.Wrapf(err, "postgres: query scan failed query=%q", query)
	}
	return &c, nil
}

// IsCartExists returns true if the cart with the given UUID exists.
func (m *PgModel) IsCartExists(ctx context.Context, cartUUID string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM cart WHERE uuid=$1) AS exists`
	var exists bool
	if err := m.db.QueryRowContext(ctx, query, cartUUID).Scan(&exists); err != nil {
		return false, errors.Wrapf(err, "postgres: query=%q scan failed", query)
	}
	if exists == true {
		return true, nil
	}
	return false, nil
}

// AddProductToCart adds a new item with productUUID and qty.
func (m *PgModel) AddProductToCart(ctx context.Context, cartUUID, userUUID, productUUID string, qty int) (*CartProductJoinRow, error) {
	tx, err := m.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, errors.Wrap(err, "postgres: db.BeginTx")
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
		return nil, errors.Wrapf(err, "postgres: query row context failed for q1=%q", q1)
	}

	q2 := "SELECT id FROM usr WHERE uuid = $1"
	var userID int
	err = tx.QueryRowContext(ctx, q2, userUUID).Scan(&userID)
	if err != nil {
		if err == sql.ErrNoRows {
			tx.Rollback()
			return nil, ErrUserNotFound
		}
		tx.Rollback()
		return nil, errors.Wrapf(err, "postgres: query row context failed for q2=%q", q2)
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
		return nil, errors.Wrapf(err, "postgres: query row context failed for q3=%q", q3)
	}

	var priceListID int
	if userUUID != "" {
		q4 := "SELECT price_list_id FROM usr WHERE uuid = $1"
		err = tx.QueryRowContext(ctx, q4, userUUID).Scan(&priceListID)
		if err != nil {
			if err == sql.ErrNoRows {
				tx.Rollback()
				return nil, ErrUserNotFound
			}
			tx.Rollback()
			return nil, errors.Wrapf(err, "postgres: query row context failed for q4=%q", q4)
		}
	} else {
		q4 := "SELECT id FROM price_list WHERE code = 'default'"
		err = tx.QueryRowContext(ctx, q4).Scan(&priceListID)
		if err != nil {
			if err == sql.ErrNoRows {
				tx.Rollback()
				return nil, ErrDefaultPriceListNotFound
			}
			tx.Rollback()
			return nil, errors.Wrapf(err, "postgres: query row context failed for q4=%q", q4)
		}
	}

	// check if the product is already in the cart
	q5 := `
		SELECT EXISTS(
		  SELECT 1 FROM cart_product WHERE cart_id = $1 AND product_id = $2
		) AS exists`
	var exists bool
	tx.QueryRowContext(ctx, q5, cartID, productID).Scan(&exists)
	if exists == true {
		tx.Rollback()
		return nil, ErrCartProductExists
	}

	// ensure there is at least one price for this product / price list combo.
	q6 := `
		SELECT COUNT(*) AS count
		FROM price
		WHERE product_id = $1 AND price_list_id = $2
	`
	var priceCount int
	row := tx.QueryRowContext(ctx, q6, productID, priceListID)
	if err := row.Scan(&priceCount); err != nil {
		tx.Rollback()
		return nil, errors.Wrapf(err, "postgres: query scan failed q6=%q", q6)
	}

	if priceCount < 1 {
		tx.Rollback()
		return nil, ErrProductHasNoPrices
	}

	q7 := `
		INSERT INTO cart_product
		  (cart_id, product_id, qty)
		VALUES
		  ($1, $2, $3)
		RETURNING
		  id
	`
	var cartProductID int
	row = tx.QueryRowContext(ctx, q7, cartID, productID, qty)
	if err := row.Scan(&cartProductID); err != nil {
		tx.Rollback()
		return nil, errors.Wrapf(err, "postgres: query scan failed q7=%q", q7)
	}

	q8 := `
		SELECT
		  c.id, c.uuid, c.product_id, p.uuid, sku, name, qty, unit_price, c.created, c.modified
		FROM cart_product AS c
		INNER JOIN product AS p
		  ON p.id = c.product_id
		LEFT OUTER JOIN price AS r
		  ON r.product_id = p.id
		WHERE
		  c.id = $1 AND r.price_list_id = $2
		ORDER BY created ASC
	`
	item := CartProductJoinRow{}
	row = tx.QueryRowContext(ctx, q8, cartProductID, priceListID)
	if err := row.Scan(&item.id, &item.UUID, &item.productID, &item.ProductUUID, &item.SKU, &item.Name,
		&item.Qty, &item.UnitPrice, &item.Created, &item.Modified); err != nil {
		tx.Rollback()
		return nil, errors.Wrapf(err, "postgres: query scan failed q8=%q", q8)
	}
	item.CartUUID = cartUUID

	if err = tx.Commit(); err != nil {
		return nil, errors.Wrap(err, "postgres: tx.Commit")
	}
	return &item, nil
}

// HasCartProducts returns true if any cart product has previously been added
// for a given cart.
func (m *PgModel) HasCartProducts(ctx context.Context, cartUUID string) (bool, error) {
	query := "SELECT COUNT(*) AS count FROM cart_product WHERE cart_id = (SELECT id FROM cart WHERE uuid = $1)"
	var count int
	err := m.db.QueryRowContext(ctx, query, cartUUID).Scan(&count)
	if err != nil {
		return false, errors.Wrapf(err, "postgres: query row context scan query=%q", query)
	}
	if count == 0 {
		return false, nil
	}
	return true, nil
}

// GetCartProducts gets all items in the cart
func (m *PgModel) GetCartProducts(ctx context.Context, cartUUID, userUUID string) ([]*CartProductJoinRow, error) {
	tx, err := m.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, errors.Wrap(err, "postgres: db.BeginTx")
	}

	// 1. Check the cart exists.
	q1 := "SELECT id FROM cart WHERE uuid = $1"
	var cartID int
	err = tx.QueryRowContext(ctx, q1, cartUUID).Scan(&cartID)
	if err != nil {
		if err == sql.ErrNoRows {
			tx.Rollback()
			return nil, ErrCartNotFound
		}
		tx.Rollback()
		return nil, errors.Wrapf(err, "postgres: query row context failed for query=%q", q1)
	}

	// 2. Check the user exists.
	q2 := "SELECT id FROM usr WHERE uuid = $1"
	var userID int
	err = tx.QueryRowContext(ctx, q2, userUUID).Scan(&userID)
	if err != nil {
		if err == sql.ErrNoRows {
			tx.Rollback()
			return nil, ErrUserNotFound
		}
		tx.Rollback()
		return nil, errors.Wrapf(err, "postgres: query row context failed for query=%q", q2)
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
	} else {
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
	q5 := `
		SELECT
		  c.id, c.uuid, c.product_id, p.uuid as product_uuid, sku, name, c.qty, unit_price, c.created, c.modified
		FROM cart_product AS c
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

	cartItems := make([]*CartProductJoinRow, 0, 20)
	for rows.Next() {
		c := CartProductJoinRow{}
		if err = rows.Scan(&c.id, &c.UUID, &c.productID, &c.ProductUUID, &c.SKU, &c.Name, &c.Qty, &c.UnitPrice, &c.Created, &c.Modified); err != nil {
			return nil, errors.Wrapf(err, "scan cart item %v", c)
		}
		c.CartUUID = cartUUID
		cartItems = append(cartItems, &c)
	}
	if err = rows.Err(); err != nil {
		return nil, errors.Wrapf(err, "postgres: rows err")
	}

	if err = tx.Commit(); err != nil {
		return nil, errors.Wrap(err, "postgres: tx.Commit")
	}
	return cartItems, nil
}

// UpdateCartProduct updates the qty of a cart product of the given product id.
func (m *PgModel) UpdateCartProduct(ctx context.Context, userUUID, cartProductUUID string, qty int) (*CartProductJoinRow, error) {
	tx, err := m.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, errors.Wrap(err, "postgres: db.BeginTx")
	}

	q1 := "SELECT id FROM cart_product WHERE uuid = $1"
	var cartProductID int
	err = tx.QueryRowContext(ctx, q1, cartProductUUID).Scan(&cartProductID)
	if err != nil {
		if err == sql.ErrNoRows {
			tx.Rollback()
			return nil, ErrCartProductNotFound
		}
		tx.Rollback()
		return nil, errors.Wrapf(err, "postgres: query row context failed for q1=%q", q1)
	}

	var priceListID int
	if userUUID != "" {
		q2 := "SELECT price_list_id FROM usr WHERE uuid = $1"
		err = tx.QueryRowContext(ctx, q2, userUUID).Scan(&priceListID)
		if err != nil {
			if err == sql.ErrNoRows {
				tx.Rollback()
				return nil, ErrUserNotFound
			}
			tx.Rollback()
			return nil, errors.Wrapf(err, "postgres: query row context failed for q2=%q", q2)
		}
	} else {
		q2 := "SELECT id FROM price_list WHERE code = 'default'"
		err = tx.QueryRowContext(ctx, q2).Scan(&priceListID)
		if err != nil {
			if err == sql.ErrNoRows {
				tx.Rollback()
				return nil, ErrDefaultPriceListNotFound
			}
			tx.Rollback()
			return nil, errors.Wrapf(err, "postgres: query row context failed for q2=%q", q2)
		}
	}

	q3 := `
		UPDATE cart_product
		SET
		  qty = $1, modified = NOW()
		WHERE
		  id = $2
		RETURNING id
	`
	var cartItemID int
	row := tx.QueryRowContext(ctx, q3, qty, cartProductID)
	if err := row.Scan(&cartItemID); err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrCartProductNotFound
		}
		tx.Rollback()
		return nil, errors.Wrapf(err, "postgres: query row context failed for q3=%q", q3)
	}

	q4 := `
		SELECT
		  c.id, c.uuid, c.product_id, p.uuid, sku, name, qty, unit_price, c.created, c.modified
		FROM cart_product AS c
		INNER JOIN product AS p
		  ON p.id = c.product_id
		LEFT OUTER JOIN price AS r
		  ON r.product_id = p.id
		WHERE
		  c.id = $1 AND r.price_list_id = $2
		ORDER BY created ASC
	`
	item := CartProductJoinRow{}
	row = tx.QueryRowContext(ctx, q4, cartItemID, priceListID)
	if err := row.Scan(&item.id, &item.UUID, &item.productID, &item.ProductUUID, &item.SKU, &item.Name,
		&item.Qty, &item.UnitPrice, &item.Created, &item.Modified); err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrCartProductNotFound
		}
		return nil, errors.Wrapf(err, "postgres: query row scan failed for q4=%q", q4)
	}

	if err = tx.Commit(); err != nil {
		return nil, errors.Wrap(err, "postgres: tx.Commit")
	}
	return &item, nil
}

// DeleteCartProduct deletes a single cart item. If the cart product
// is not in the cart it returns `ErrCartProductNotFound`.
func (m *PgModel) DeleteCartProduct(ctx context.Context, cartProductUUID string) error {
	tx, err := m.db.BeginTx(ctx, nil)
	if err != nil {
		return errors.Wrap(err, "postgres: db.BeginTx failed")
	}

	q1 := "SELECT id FROM cart_product WHERE uuid = $1"
	var cartProductID int
	if err = tx.QueryRowContext(ctx, q1, cartProductUUID).Scan(&cartProductID); err != nil {
		if err == sql.ErrNoRows {
			tx.Rollback()
			return ErrCartNotFound
		}
		tx.Rollback()
		return errors.Wrapf(err, "postgres: query row context failed for q1=%q", q1)
	}

	q2 := "DELETE FROM cart_product WHERE id = $1"
	_, err = tx.ExecContext(ctx, q2, cartProductID)
	if err != nil {
		return errors.Wrapf(err, "postgres: exec context failed q2=%q", q2)
	}

	if err = tx.Commit(); err != nil {
		return errors.Wrap(err, "postgres: tx.Commit failed")
	}

	return nil
}

// EmptyCartProducts empty the cart of all items. Does not affect coupons.
func (m *PgModel) EmptyCartProducts(ctx context.Context, cartUUID string) (err error) {
	tx, err := m.db.BeginTx(ctx, nil)
	if err != nil {
		return errors.Wrap(err, "postgres: db.BeginTx failed")
	}

	// 1. Check the cart exists
	q1 := "SELECT id FROM cart WHERE uuid = $1"
	var cartID int
	err = tx.QueryRowContext(ctx, q1, cartUUID).Scan(&cartID)
	if err != nil {
		if err == sql.ErrNoRows {
			tx.Rollback()
			return ErrCartNotFound
		}
		tx.Rollback()
		return errors.Wrapf(err, "postgres: query row context failed for q1=%q", q1)
	}

	// 2. Delete all product from the cart
	q2 := "DELETE FROM cart_product WHERE cart_id = $1"
	_, err = m.db.ExecContext(ctx, q2, cartID)
	if err != nil {
		return errors.Wrapf(err, "postgres: exec context q2=%q", q2)
	}

	if err = tx.Commit(); err != nil {
		return errors.Wrap(err, "postgres: tx.Commit failed")
	}
	return nil
}
