package model

import (
	"context"
	"time"

	"bitbucket.org/andyfusniakteam/ecom-api-go/utils/nestedset"
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
	CatalogModel
}

type CartModel interface {
	CreateCart(ctx context.Context) (*string, error)
	AddItemToCart(ctx context.Context, cartUUID, tierRef, sku string, qty int) (*CartItem, error)
	GetCartItems(ctx context.Context, cartUUID string) ([]*CartItem, error)
	UpdateItemByCartUUID(ctx context.Context, cartUUID, sku string, qty int) (*CartItem, error)
	DeleteCartItem(ctx context.Context, cartUUID, sku string) (count int64, err error)
	EmptyCartItems(ctx context.Context, cartUUID string) (err error)
}

type CustomerModel interface {
	CreateCustomer(ctx context.Context, UID, email, firstname, lastname string) (*Customer, error)
	GetCustomers(ctx context.Context, page, size int, startsAfter string) ([]*Customer, error)
	GetCustomerByUUID(ctx context.Context, customerUUID string) (*Customer, error)
	GetCustomerIDByUUID(ctx context.Context, customerUUID string) (int, error)
}

type AddressModel interface {
	CreateAddress(ctx context.Context, customerID int, typ, contactName, addr1 string, addr2 *string, city string, county *string, postcode, country string) (*Address, error)
	GetAddressByUUID(ctx context.Context, addrUUID string) (*Address, error)
	GetAddressOwnerByUUID(ctx context.Context, addrUUID string) (*string, error)
	GetAddresses(ctx context.Context, customerID int) ([]*Address, error)
	UpdateAddressByUUID(ctx context.Context, addrUUID string) (*Address, error)
	DeleteAddressByUUID(ctx context.Context, addrUUID string) error
}

// CatalogModel interface
type CatalogModel interface {
	GetCatalogNestedSet(ctx context.Context) ([]*nestedset.NestedSetNode, error)
}
