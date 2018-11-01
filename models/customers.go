package models

import (
	"database/sql"
	"time"
)

// DB is a database handle
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
	Typ         string         `json:"type"`
	ContactName string         `json:"contact_name"`
	Addr1       string         `json:"address_line_1"`
	Addr2       sql.NullString `json:"address_line_2"`
	City        string         `json:"city"`
	County      sql.NullString `json:"county"`
	Postcode    string         `json:"postcode"`
	Country     string         `json:"country"`
	Created     time.Time      `json:"created"`
	Modified    time.Time      `json:"modified"`
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
func CreateCustomer(firstname string, lastname string) (*Customer, error) {
	c := Customer{}

	sql := `INSERT INTO customers (firstname, lastname) VALUES ($1, $2) RETURNING id, customer_uuid, firstname, lastname, created, modified`
	err := DB.QueryRow(sql, firstname, lastname).Scan(
		&c.id, &c.CustomerUUID, &c.Firstname, &c.Lastname, &c.Created, &c.Modified)
	if err != nil {
		panic(err)
	}

	return &c, nil
}

// GetCustomerByUUID gets a customer by customer UUID
func GetCustomerByUUID(customerUUID string) *Customer {
	c := Customer{}
	sql := `
	SELECT
		id, customer_uuid, firstname, lastname, created, modified
	FROM customers
	WHERE customer_uuid = $1`
	err := DB.QueryRow(sql, customerUUID).Scan(&c.id, &c.CustomerUUID, &c.Firstname, &c.Lastname, &c.Created, &c.Modified)
	if err != nil {
		panic(err)
	}
	return &c
}

// CreateAddress creates a new billing or shipping address for a customer
func CreateAddress(customerID int, typ string, contactName string, addr1 string,
	addr2 string, city string, county string, postcode string, country string) *Address {
	a := Address{}

	sql := `
	INSERT INTO addresses (
		customer_id, typ, contact_name, addr1, addr2, city, county, postcode, country 
	) VALUES (
		$1, $2, $3, $4, $5, $6, $7, $8, $9
	) RETURNING
		id, addr_uuid, customer_id, typ, contact_name, addr1, addr2, city, county, postcode, country, created, modified
	`

	err := DB.QueryRow(sql, customerID, typ, contactName, addr1, addr2, city, county, postcode, country).Scan(&a.customerID, &a.AddrUUID, &a.customerID, &a.Typ, &a.ContactName, &a.Addr1, &a.Addr2, &a.City, &a.County, &a.Postcode, &a.Country, &a.Created, &a.Modified)
	if err != nil {
		panic(err)
	}

	return &a
}

// GetAddressByUUID gets an address by UUID. Returns a pointer to an Address.
func GetAddressByUUID(addrUUID string) *Address {
	a := Address{}

	sql := `
	SELECT
		id, address_uuid, customer_id, typ, contact_name, addr1, addr2,
		city, county, postcode, country, created, modified
	FROM addresses
	WHERE address_uuid = $1
	`

	DB.QueryRow(sql, addrUUID).Scan(&a.id, &a.customerID, &a.Typ, &a.ContactName, &a.Addr1, &a.Addr2, &a.City, &a.County, &a.Postcode, &a.Country, &a.Created, &a.Modified)

	return &a
}

// GetCustomerIDByUUID converts between customer UUID and the underlying primary key
func GetCustomerIDByUUID(customerUUID string) (int, error) {
	var id int

	query := `SELECT id FROM customers WHERE customer_uuid = $1`
	row := DB.QueryRow(query, customerUUID)
	err := row.Scan(&id)
	if err != nil {
		return 0, err
	}

	return id, nil
}

// GetAddresses retrieves a slice of addresses for a given customer
func GetAddresses(customerID int) ([]Address, error) {
	addresses := make([]Address, 0, 8)

	sql := `
	SELECT
		id, addr_uuid, customer_id, typ, contact_name, addr1, addr2, city, county,
		postcode, country, created, modified
	FROM addresses
	WHERE customer_id = $1
	`
	rows, err := DB.Query(sql, customerID)
	if err != nil {
		panic(err)
	}
	defer rows.Close()

	for rows.Next() {
		var a Address

		err = rows.Scan(&a.id, &a.AddrUUID, &a.customerID, &a.Typ, &a.ContactName, &a.Addr1, &a.Addr2, &a.City, &a.County, &a.Postcode, &a.Country, &a.Created, &a.Modified)
		if err != nil {
			return nil, err
		}

		addresses = append(addresses, a)
	}

	return addresses, nil
}

// UpdateAddressByUUID updates an address for a given customer
func UpdateAddressByUUID() *Address {

	//query := `UPDATE addresses SET`
	addr := Address{}
	return &addr
}

// DeleteAddressByUUID deletes an address by uuid
func DeleteAddressByUUID(addrUUID string) error {
	query := `DELETE FROM addresses WHERE addr_uuid = $1`

	_, err := DB.Exec(query, addrUUID)
	if err != nil {
		panic(err)
		return err
	}

	return nil
}
