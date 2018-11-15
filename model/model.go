package model

import (
	"time"
)

// CartItem structure holds the details individual cart item
type CartItem struct {
	ID        int
	CartUUID  string
	Sku       string
	Qty       int
	UnitPrice float64
	Created   time.Time
	Modified  time.Time
}

// Customer details
type Customer struct {
	ID           int
	CustomerUUID string
	UID          string
	Email        string
	Firstname    string
	Lastname     string
	Created      time.Time
	Modified     time.Time
}

// Address contains address information for a Customer
type Address struct {
	ID          int
	AddrUUID    string
	CustomerID  int
	Typ         string
	ContactName string
	Addr1       string
	Addr2       *string
	City        string
	County      *string
	Postcode    string
	Country     string
	Created     time.Time
	Modified    time.Time
}

type EcomModel interface {
	CartModel
	CustomerModel
	AddressModel
}

type CartModel interface {
	CreateCart() (*string, error)
	AddItemToCart(cartUUID, tierRef, sku string, qty int) (*CartItem, error)
	GetCartItems(cartUUID string) ([]*CartItem, error)
	UpdateItemByCartUUID(cartUUID, sku string, qty int) (*CartItem, error)
	DeleteCartItem(cartUUID, sku string) (count int64, err error)
	EmptyCartItems(cartUUID string) (err error)
}

type CustomerModel interface {
	CreateCustomer(UID, email, firstname, lastname string) (*Customer, error)
	GetCustomerByUUID(customerUUID string) (*Customer, error)
	GetCustomerIDByUUID(customerUUID string) (int, error)
}

type AddressModel interface {
	CreateAddress(customerID int, typ, contactName, addr1 string, addr2 *string, city string, county *string, postcode, country string) (*Address, error)
	GetAddressByUUID(addrUUID string) (*Address, error)
	GetAddresses(customerID int) ([]*Address, error)
	UpdateAddressByUUID(addrUUID string) (*Address, error)
	DeleteAddressByUUID(addrUUID string) error
}
