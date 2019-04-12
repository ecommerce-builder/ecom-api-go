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
	Role         string
	Email        string
	Firstname    string
	Lastname     string
	Created      time.Time
	Modified     time.Time
}

// PaginationResultSet contains both the underlying result set as well as
// context about the data including Total; the total number of rows in
// the table, First; set to true if this result set represents the first
// page, Last; set to true if this result set represents the last page of
// results.
type PaginationResultSet struct {
	RContext struct {
		Total               int
		FirstUUID, LastUUID string
	}
	RSet interface{}
}

type PaginationQuery struct {
	OrderBy    string
	OrderDir   string
	Limit      int
	StartAfter string
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

type CustomerDevKey struct {
	ID           int       `json:"id"`
	UUID         string    `json:"uuid"`
	Key          string    `json:"key"`
	Hash         string    `json:"hash"`
	CustomerID   int       `json:"customer_id"`
	CustomerUUID string    `json:"customer_uuid"`
	Created      time.Time `json:"created"`
	Modified     time.Time `json:"modified"`
}

type (
	ProductUpdate struct {
		EAN  string `json:"ean"`
		URL  string `json:"url"`
		Name string `json:"name"`
	}

	Product struct {
		ID       int       `json:"id"`
		UUID     string    `json:"uuid"`
		SKU      string    `json:"sku"`
		EAN      string    `json:"ean"`
		URL      string    `json:"url"`
		Name     string    `json:"name"`
		Created  time.Time `json:"created"`
		Modified time.Time `json"modified"`
	}
)

// CatalogProductAssoc maps products to leaf nodes in the catalogue hierarchy
type CatalogProductAssoc struct {
	ID        int
	CatalogID int
	ProductID int
	Path      string `json:"path"`
	SKU       string `json:"sku"`
	Pri       int    `json:"pri"`
	Created   time.Time
	Modified  time.Time
}

type EcomModel interface {
	CartModel
	CustomerModel
	AddressModel
	ProductModel
	CatalogModel
	ErrorModel
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
	CreateCustomer(ctx context.Context, uid, role, email, firstname, lastname string) (*Customer, error)
	GetCustomers(ctx context.Context, q *PaginationQuery) (*PaginationResultSet, error)
	GetCustomerByUUID(ctx context.Context, customerUUID string) (*Customer, error)
	GetCustomerByID(ctx context.Context, customerID int) (*Customer, error)
	GetCustomerIDByUUID(ctx context.Context, customerUUID string) (int, error)
	CreateCustomerDevKey(ctx context.Context, customerID int, apiKey string) (*CustomerDevKey, error)
	GetCustomerDevKey(ctx context.Context, uuid string) (*CustomerDevKey, error)
	GetCustomerDevKeys(ctx context.Context, customerID int) ([]*CustomerDevKey, error)
	GetCustomerDevKeyByDevKey(ctx context.Context, devKey string) (*CustomerDevKey, error)
}

type AddressModel interface {
	CreateAddress(ctx context.Context, customerID int, typ, contactName, addr1 string, addr2 *string, city string, county *string, postcode, country string) (*Address, error)
	GetAddressByUUID(ctx context.Context, addrUUID string) (*Address, error)
	GetAddressOwnerByUUID(ctx context.Context, addrUUID string) (*string, error)
	GetAddresses(ctx context.Context, customerID int) ([]*Address, error)
	UpdateAddressByUUID(ctx context.Context, addrUUID string) (*Address, error)
	DeleteAddressByUUID(ctx context.Context, addrUUID string) error
}

type ProductModel interface {
	CreateProduct(ctx context.Context, sku string) (*Product, error)
	GetProduct(ctx context.Context, sku string) (*Product, error)
	UpdateProduct(ctx context.Context, sku string, pu *ProductUpdate) (*Product, error)
	DeleteProduct(ctx context.Context, sku string) error
}

// CatalogModel interface
type CatalogModel interface {
	GetCatalogNestedSet(ctx context.Context) ([]*nestedset.NestedSetNode, error)
	GetCatalogProductAssocs(ctx context.Context) ([]*CatalogProductAssoc, error)
	UpdateCatalogProductAssocs(ctx context.Context, cpo []*CatalogProductAssoc) error
}

type CreateProductImage struct {
	SKU   string
	W     uint
	H     uint
	Path  string
	Typ   string
	Ori   bool
	Pri   uint
	Size  uint
	Q     uint
	GSURL string
	Data  interface{}
}
type ProductImage struct {
	ID        uint
	ProductID uint
	UUID      string
	SKU       string
	W         uint
	H         uint
	Path      string
	Typ       string
	Ori       bool
	Up        bool
	Pri       uint
	Size      uint
	Q         uint
	GSURL     string
	Data      interface{}
	Created   time.Time
	Modified  time.Time
}

type ProductImagesModel interface {
	CreateImageEntry(ctx context.Context, p *CreateProductImage) (*ProductImage, error)
	ConfirmImageUploaded(ctx context.Context, uuid string) (*ProductImage, error)
}

type ErrorModel interface {
	IsNotExist(err error) bool
}
