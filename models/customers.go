package models

import (
	"database/sql"
	"time"
)

var DB *sql.DB

// Customer details
type Customer struct {
	id           int
	CustomerUUID string    `json:"customer_uuid"`
	Firstname    string    `json:"firstname"`
	Lastname     string    `json:"lastname"`
	Created      time.Time `json:"created"`
	Modified     time.Time `json:"modified"`
}

// Address contains address information for a Customer
type Address struct {
	id          int
	AddrUUID    string `json:"address_uuid"`
	customerID  int
	Typ         string    `json:"type"`
	ContactName string    `json:"contact_name"`
	Addr1       string    `json:"address_line_1"`
	Addr2       string    `json:"address_line_2"`
	City        string    `json:"city"`
	County      string    `json:"county"`
	Country     string    `json:"country"`
	Postcode    string    `json:"postcode"`
	Created     time.Time `json:"created"`
	Modified    time.Time `json:"modified"`
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
func CreateCustomer(firstname string, lastname string) *Customer {
	c := Customer{}

	sql := `INSERT INTO customers (firstname, lastname) VALUES ($1, $2) RETURNING *`
	err := DB.QueryRow(sql, firstname, lastname).Scan(
		&c.id, &c.CustomerUUID, &c.Firstname, &c.Lastname, &c.Created, &c.Modified)
	if err != nil {
		panic(err)
	}
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

	err := DB.QueryRow(sql, c.id, typ, contactName, addr1, addr2, city, county, country, postcode).Scan(&a.customerID, &a.AddrUUID, &a.customerID, &a.Typ, &a.ContactName, &a.Addr1, &a.Addr2, &a.City, &a.County, &a.Country, &a.Postcode, &a.Created, &a.Modified)
	if err != nil {
		panic(err)
	}

	return &a
}

// GetAddress returns a pointer to an Address structure for a specific customer given the id
func (c *Customer) GetAddress(id int) *Address {
	a := Address{}

	query := `
	SELECT
		id, address_uuid, customer_id, typ, contact_name, addr_line_1, addr_line_2,
		city, county, country, postcode, created, modified
	FROM addresses
	WHERE customer_id = $1 AND id = $2
	`

	DB.QueryRow(query, c.CustomerUUID, id).Scan(&a.id, &a.customerID, &a.Typ, &a.ContactName, &a.Addr1, &a.Addr2, &a.City, &a.County, &a.Country, &a.Postcode, &a.Created, &a.Modified)

	return &a
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
