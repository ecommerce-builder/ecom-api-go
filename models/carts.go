package models

import (
	"fmt"
	"strconv"
	"time"
)

// CartItem structure holds the details individual cart ite
type CartItem struct {
	id        int
	CartUUID  string    `json:"cart_uuid"`
	Sku       string    `json:"sku"`
	Qty       int       `json:"qty"`
	UnitPrice float64   `json:"unit_price"`
	Created   time.Time `json:"created"`
	Modified  time.Time `json:"modified"`
}

// CreateCart creates a new shopping cart
func CreateCart() string {
	var cartUUID string

	query := `SELECT UUID_GENERATE_V4() AS cart_uuid`

	err := DB.QueryRow(query).Scan(&cartUUID)
	if err != nil {
		panic(err)
	}

	return cartUUID
}

// AddItemToCart adds a new item with sku, qty and unit price
func AddItemToCart(tierRef string, cartUUID string, sku string, qty int) (*CartItem, error) {
	item := CartItem{}

	// check if the item is alreadyd in the cart
	query := `SELECT EXISTS(SELECT 1 FROM carts WHERE cart_uuid=$1 AND sku=$2) AS exists;`
	var exists bool
	DB.QueryRow(query, cartUUID, sku).Scan(&exists)
	if exists == true {
		return nil, fmt.Errorf("Cart item %s already exists", sku)
	}

	var unitPriceStr []byte
	query = `SELECT unit_price FROM product_pricing WHERE tier_ref = $1 AND sku = $2`
	err := DB.QueryRow(query, tierRef, sku).Scan(&unitPriceStr)
	if err != nil {
		return &item, err
	}

	unitPrice, _ := strconv.ParseFloat(string(unitPriceStr), 64)
	query = `
	INSERT INTO carts (cart_uuid, sku, qty, unit_price) VALUES ($1, $2, $3, $4)
	RETURNING id, cart_uuid, sku, qty, unit_price, created, modified
	`

	err = DB.QueryRow(query, cartUUID, sku, qty, unitPrice).Scan(&item.id, &item.CartUUID, &item.Sku, &item.Qty, &item.UnitPrice, &item.Created, &item.Modified)
	if err != nil {
		panic(err)
		return nil, err
	}

	return &item, nil
}

// GetCartItems get all items in the given cart
func GetCartItems(cartUUID string) ([]CartItem, error) {
	cartItems := make([]CartItem, 0, 20)

	sql := `
	SELECT
		id, cart_uuid, sku, qty, unit_price, created, modified
	FROM carts
	WHERE cart_uuid = $1
	`
	rows, err := DB.Query(sql, cartUUID)
	if err != nil {
		panic(err)
	}
	defer rows.Close()

	for rows.Next() {
		var c CartItem

		err = rows.Scan(&c.id, &c.CartUUID, &c.Sku, &c.Qty, &c.UnitPrice, &c.Created, &c.Modified)
		if err != nil {
			return nil, err
		}

		cartItems = append(cartItems, c)

	}

	return cartItems, nil
}

// UpdateCartItemByUUID updates the qty of a cart item of the given sku.
func UpdateCartItemByUUID(cartUUID string, sku string, qty int) (*CartItem, error) {
	query := `
	UPDATE carts
	SET qty = $1, modified = NOW()
	WHERE cart_uuid = $2 AND sku = $3
	RETURNING id, cart_uuid, sku, qty, unit_price, created, modified
	`

	item := CartItem{}
	err := DB.QueryRow(query, qty, cartUUID, sku).Scan(&item.id, &item.CartUUID, &item.Sku, &item.Qty, &item.UnitPrice, &item.Created, &item.Modified)
	if err != nil {
		panic(err)
		return &item, err
	}

	return &item, nil
}

// DeleteCartItem deletes a single cart item for a given card.
func DeleteCartItem(cartUUID string, sku string) (count int64, err error) {
	query := `DELETE FROM carts WHERE cart_uuid = $1 AND sku = $2`
	res, err := DB.Exec(query, cartUUID, sku)
	if err != nil {
		panic(err)
		return 0, err
	}
	count, err = res.RowsAffected()
	if err != nil {
		panic(err)
		return 0, err
	}

	return count, nil
}

// EmptyCartItems empty the cart of all items. Does not affect coupons.
func EmptyCartItems(cartUUID string) (err error) {
	query := `DELETE FROM carts WHERE cart_uuid = $1`
	_, err = DB.Exec(query, cartUUID)
	if err != nil {
		panic(err)
		return err
	}

	return nil
}
