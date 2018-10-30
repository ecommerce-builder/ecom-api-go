package models

import (
	"database/sql"
	"time"
)

// Customer details
type Customer struct {
	db           *sql.DB
	id           int
	customerUUID string
	firstname    string
	lastname     string
	created      time.Time
	modified     time.Time
}

// Address contains address information for a Customer
type Address struct {
	id          int
	addrUUID    string
	customerID  int
	typ         string
	contactName string
	addr1       string
	addr2       string
	city        string
	county      string
	country     string
	postcode    string
	created     time.Time
	modified    time.Time
}

// CustomerModeler interface
type CustomerModeler interface {
	CreateCustomer(firstname string, lastname string) *Customer
	GetCustomer(id int) *Customer
}

// AddressModeler interface
type AddressModeler interface {
	CreateAddress(typ string, contactName string, addr1 string, addr2 string, city string, county string, country string, postcode string) *Address
	GetAddress(id int) *Address
	ListAddresses() []Address
	UpdateAddress() *Address
	DeleteAddress()
}

// CreateCustomer creates a new customer
func CreateCustomer(db *sql.DB, firstname string, lastname string) *Customer {
	c := Customer{}

	sql := `INSERT INTO customers (firstname, lastname) VALUES ($1, $2) RETURNING *`

	err := db.QueryRow(sql, firstname, lastname).Scan(
		&c.id, &c.customerUUID, &c.firstname, &c.lastname, &c.created, &c.modified)
	if err != nil {
		panic(err)
	}

	c.db = db
	return &c
}

// CreateAddress creates a new billing or shipping address for a customer
func (c *Customer) CreateAddress(typ string, contactName string, addr1 string, addr2 string, city string, county string, country string, postcode string) *Address {
	a := Address{}

	sql := `
	INSERT INTO addresses (
		customer_id, typ, contact_name, addr1, addr2, city, county, country, postcode
	) VALUES (
		$1, $2, $3, $4, $5, $6, $7, $8, $9
	) RETURNING
		id, addr_uuid, customer_id, typ, contact_name, addr1, addr2, city, county, country, postcode, created, modified
	`

	err := c.db.QueryRow(sql, c.id, typ, contactName, addr1, addr2, city, county, country, postcode).Scan(&a.customerID, &a.addrUUID, &a.customerID, &a.typ, &a.contactName, &a.addr1, &a.addr2, &a.city, &a.county, &a.country, &a.postcode, &a.created, &a.modified)
	if err != nil {
		panic(err)
	}

	return &a
}

// GetAddress returns a pointer to an Address structure for a specific customer given the id
func (c *Customer) GetAddress(id int) *Address {
	addr := Address{}

	query := `
	SELECT
		id, address_uuid, customer_id, typ, contact_name, addr_line_1, addr_line_2,
		city, county, country, postcode, created, modified
	FROM addresses
	WHERE customer_id = $1 AND id = $2
	`

	c.db.QueryRow(query, c.customerUUID, id).Scan(&addr.id, &addr.customerID, &addr.typ, &addr.contactName, &addr.addr1, &addr.addr2, &addr.city, &addr.county, &addr.country, &addr.postcode, &addr.created, &addr.modified)

	return &addr
}

// ListAddresses retrieves a slice of Address for a given customer
func (c *Customer) ListAddresses() []Address {
	var l = []Address{}

	return l
}

// UpdateAddress updates an address for a given customer
func (c *Customer) UpdateAddress() *Address {
	addr := Address{}
	return &addr
}
