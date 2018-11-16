package postgres

import (
	"database/sql"
	"fmt"
	"strconv"

	"bitbucket.org/andyfusniakteam/ecom-api-go/model"
)

type PgModel struct {
	db *sql.DB
}

func New(db *sql.DB) (*PgModel, error) {
	pgModel := PgModel{db}
	return &pgModel, nil
}

// CreateCart creates a new shopping cart
func (m *PgModel) CreateCart() (*string, error) {
	var cartUUID string

	query := `SELECT UUID_GENERATE_V4() AS cart_uuid`
	err := m.db.QueryRow(query).Scan(&cartUUID)
	if err != nil {
		return nil, err
	}

	return &cartUUID, nil
}

// AddItemToCart adds a new item with sku, qty and unit price
func (m *PgModel) AddItemToCart(cartUUID, tierRef, sku string, qty int) (*model.CartItem, error) {
	item := model.CartItem{}

	// check if the item is alreadyd in the cart
	query := `SELECT EXISTS(SELECT 1 FROM carts WHERE cart_uuid=$1 AND sku=$2) AS exists;`
	var exists bool
	m.db.QueryRow(query, cartUUID, sku).Scan(&exists)
	if exists == true {
		return nil, fmt.Errorf("Cart item %s already exists", sku)
	}

	var unitPriceStr []byte
	query = `SELECT unit_price FROM product_pricing WHERE tier_ref = $1 AND sku = $2`
	err := m.db.QueryRow(query, tierRef, sku).Scan(&unitPriceStr)
	if err != nil {
		return &item, err
	}

	unitPrice, _ := strconv.ParseFloat(string(unitPriceStr), 64)
	query = `
		INSERT INTO carts (cart_uuid, sku, qty, unit_price)
		VALUES ($1, $2, $3, $4)
		RETURNING id, cart_uuid, sku, qty, unit_price, created, modified
	`

	err = m.db.QueryRow(query, cartUUID, sku, qty, unitPrice).Scan(&item.ID, &item.CartUUID, &item.Sku, &item.Qty, &item.UnitPrice, &item.Created, &item.Modified)
	if err != nil {
		return nil, err
	}

	return &item, nil
}

// GetCartItems gets all items in the cart
func (m *PgModel) GetCartItems(cartUUID string) ([]*model.CartItem, error) {
	cartItems := make([]*model.CartItem, 0, 20)

	query := `
		SELECT
			id, cart_uuid, sku, qty, unit_price, created, modified
		FROM carts
		WHERE cart_uuid = $1
	`
	rows, err := m.db.Query(query, cartUUID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		c := model.CartItem{}
		err = rows.Scan(&c.ID, &c.CartUUID, &c.Sku, &c.Qty, &c.UnitPrice, &c.Created, &c.Modified)
		if err != nil {
			return nil, err
		}

		cartItems = append(cartItems, &c)
	}

	return cartItems, nil
}

// UpdateItemByCartUUID updates the qty of a cart item of the given sku.
func (m *PgModel) UpdateItemByCartUUID(cartUUID, sku string, qty int) (*model.CartItem, error) {
	query := `
		UPDATE carts
		SET qty = $1, modified = NOW()
		WHERE cart_uuid = $2 AND sku = $3
		RETURNING id, cart_uuid, sku, qty, unit_price, created, modified
	`

	item := model.CartItem{}
	err := m.db.QueryRow(query, qty, cartUUID, sku).Scan(&item.ID, &item.CartUUID, &item.Sku, &item.Qty, &item.UnitPrice, &item.Created, &item.Modified)
	if err != nil {
		return nil, err
	}

	return &item, nil
}

// DeleteCartItem deletes a single cart item.
func (m *PgModel) DeleteCartItem(cartUUID, sku string) (count int64, err error) {
	query := `DELETE FROM carts WHERE cart_uuid = $1 AND sku = $2`
	res, err := m.db.Exec(query, cartUUID, sku)
	if err != nil {
		return -1, err
	}
	count, err = res.RowsAffected()
	if err != nil {
		return -1, err
	}

	return count, nil
}

// EmptyCartItems empty the cart of all items. Does not affect coupons.
func (m *PgModel) EmptyCartItems(cartUUID string) (err error) {
	query := `DELETE FROM carts WHERE cart_uuid = $1`
	_, err = m.db.Exec(query, cartUUID)
	if err != nil {
		return err
	}

	return nil
}

// CreateCustomer creates a new customer
func (m *PgModel) CreateCustomer(UID, email, firstname, lastname string) (*model.Customer, error) {
	c := model.Customer{}
	query := `
		INSERT INTO customers (
			uid, email, firstname, lastname
		) VALUES (
			$1, $2, $3, $4
		)
		RETURNING id, customer_uuid, uid, email, firstname, lastname, created, modified
	`
	err := m.db.QueryRow(query, UID, email, firstname, lastname).Scan(
		&c.ID, &c.CustomerUUID, &c.UID, &c.Email, &c.Firstname, &c.Lastname, &c.Created, &c.Modified)
	if err != nil {
		return nil, err
	}

	return &c, nil
}

// GetCustomerByUUID gets a customer by customer UUID
func (m *PgModel) GetCustomerByUUID(customerUUID string) (*model.Customer, error) {
	c := model.Customer{}
	query := `
		SELECT
			id, customer_uuid, uid, email, firstname, lastname, created, modified
		FROM customers
		WHERE customer_uuid = $1
	`
	err := m.db.QueryRow(query, customerUUID).Scan(&c.ID, &c.CustomerUUID, &c.UID, &c.Email, &c.Firstname, &c.Lastname, &c.Created, &c.Modified)
	if err != nil {
		return nil, err
	}

	return &c, nil
}

// CreateAddress creates a new billing or shipping address for a customer
func (m *PgModel) CreateAddress(customerID int, typ, contactName, addr1 string, addr2 *string, city string, county *string, postcode, country string) (*model.Address, error) {
	a := model.Address{}
	query := `
		INSERT INTO addresses (
			customer_id, typ, contact_name, addr1, addr2, city, county, postcode, country
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9
		) RETURNING
			id, addr_uuid, customer_id, typ, contact_name, addr1, addr2, city, county, postcode, country, created, modified
	`

	err := m.db.QueryRow(query, customerID, typ, contactName, addr1, addr2, city, county, postcode, country).Scan(&a.ID, &a.AddrUUID, &a.CustomerID, &a.Typ, &a.ContactName, &a.Addr1, &a.Addr2, &a.City, &a.County, &a.Postcode, &a.Country, &a.Created, &a.Modified)
	if err != nil {
		return nil, err
	}

	return &a, nil
}

// GetAddressByUUID gets an address by UUID. Returns a pointer to an Address.
func (m *PgModel) GetAddressByUUID(addrUUID string) (*model.Address, error) {
	a := model.Address{}
	query := `
		SELECT
			id, addr_uuid, customer_id, typ, contact_name, addr1, addr2,
			city, county, postcode, country, created, modified
		FROM addresses
		WHERE addr_uuid = $1
	`

	err := m.db.QueryRow(query, addrUUID).Scan(&a.ID, &a.AddrUUID, &a.CustomerID, &a.Typ, &a.ContactName, &a.Addr1, &a.Addr2, &a.City, &a.County, &a.Postcode, &a.Country, &a.Created, &a.Modified)
	if err != nil {
		return nil, err
	}

	return &a, nil
}

// GetCustomerIDByUUID converts between customer UUID and the underlying primary key
func (m *PgModel) GetCustomerIDByUUID(customerUUID string) (int, error) {
	var id int

	query := `SELECT id FROM customers WHERE customer_uuid = $1`
	row := m.db.QueryRow(query, customerUUID)
	err := row.Scan(&id)
	if err != nil {
		return -1, err
	}

	return id, nil
}

// GetAddresses retrieves a slice of pointers to Address for a given customer
func (m *PgModel) GetAddresses(customerID int) ([]*model.Address, error) {
	addresses := make([]*model.Address, 0, 8)

	query := `
		SELECT
			id, addr_uuid, customer_id, typ, contact_name, addr1,
			addr2, city, county, postcode, country, created, modified
		FROM addresses
		WHERE customer_id = $1
		ORDER BY created DESC
	`
	rows, err := m.db.Query(query, customerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var a model.Address

		err = rows.Scan(&a.ID, &a.AddrUUID, &a.CustomerID, &a.Typ, &a.ContactName, &a.Addr1, &a.Addr2, &a.City, &a.County, &a.Postcode, &a.Country, &a.Created, &a.Modified)
		if err != nil {
			return nil, err
		}

		addresses = append(addresses, &a)
	}

	return addresses, nil
}

// UpdateAddressByUUID updates an address for a given customer
func (m *PgModel) UpdateAddressByUUID(addrUUID string) (*model.Address, error) {
	// TO BE DONE
	//
	//query := `UPDATE addresses SET`
	addr := model.Address{}
	return &addr, nil
}

// DeleteAddressByUUID deletes an address by uuid
func (m *PgModel) DeleteAddressByUUID(addrUUID string) error {
	query := `DELETE FROM addresses WHERE addr_uuid = $1`

	_, err := m.db.Exec(query, addrUUID)
	if err != nil {
		return err
	}

	return nil
}
