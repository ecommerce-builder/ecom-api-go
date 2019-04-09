package app

import (
	"context"
	"net/http"
	"time"

	"bitbucket.org/andyfusniakteam/ecom-api-go/model"
	"bitbucket.org/andyfusniakteam/ecom-api-go/utils/nestedset"
	"firebase.google.com/go/auth"
)

const (
	OpCreateAdmin string = "CreateAdmin"
)

// Cart operation sentinel values.
const (
	OpCreateCart     string = "CreateCart"
	OpAddItemToCart  string = "AddItemToCart"
	OpGetCartItems   string = "GetCartItems"
	OpUpdateCartItem string = "UpdateCartItem"
	OpDeleteCartItem string = "DeleteCartItem"
	OpEmptyCartItems string = "EmptyCartItems"
)

const (
	OpCreateCustomer        string = "CreateCustomer"
	OpGetCustomer           string = "GetCustomer"
	OpListCustomers         string = "ListCustomers"
	OpGetCustomersAddresses string = "GetCustomersAddresses"
	OpUpdateAddress         string = "UpdateAddress"
	OpCreateAddress         string = "CreateAddress"
	OpGetAddress            string = "GetAddress"
	OpDeleteAddress         string = "DeleteAddress"
)

const (
	OpGenerateCustomerDevKey string = "GenerateCustomerDevKey"
	OpListCustomersDevKeys   string = "ListCustomersDevKeys"
	OpDeleteCustomerDevKey   string = "DeleteCustomerDevKey"
	OpSignInWithDevKey       string = "SignInWithDevKey"
)

const (
	OpGetCatalog string = "GetCatalog"
)

const (
	OpSystemInfo string = "SystemInfo"
)

const (
	RoleSuperUser string = "root"
	RoleAdmin     string = "admin"
	RoleCustomer  string = "customer"
	RoleShopper   string = "anon"
)

type Serverable interface {
	Handlerable
	EcomService
}

type EcomService interface {
	CartService
	CustomerService
	CatalogAndProductService
	AuthService
	ErrorService
}

type App struct {
	Service EcomService
}

type Handlerable interface {
	CreateCartHandler() http.HandlerFunc
	GetCartItemsHandler() http.HandlerFunc
	CreateCustomerHandler() http.HandlerFunc
	AuthenticateMiddleware(http.HandlerFunc) http.HandlerFunc
}

// CartItem structure holds the details individual cart item
type CartItem struct {
	CartUUID  string    `json:"cart_uuid"`
	Sku       string    `json:"sku"`
	Qty       int       `json:"qty"`
	UnitPrice float64   `json:"unit_price"`
	Created   time.Time `json:"created"`
	Modified  time.Time `json:"modified"`
}

type CartService interface {
	CreateCart(ctx context.Context) (*string, error)
	AddItemToCart(ctx context.Context, cartUUID, sku string, qty int) (*CartItem, error)
	GetCartItems(ctx context.Context, cartUUID string) ([]*CartItem, error)
	UpdateCartItem(ctx context.Context, cartUUID string, sku string, qty int) (*CartItem, error)
	DeleteCartItem(ctx context.Context, cartUUID string, sku string) (count int64, err error)
	EmptyCartItems(ctx context.Context, cartUUID string) (err error)
}

// Customer details
type Customer struct {
	CustomerUUID string    `json:"customer_uuid"`
	UID          string    `json:"uid"`
	Role         string    `json:"role"`
	Email        string    `json:"email"`
	Firstname    string    `json:"firstname"`
	Lastname     string    `json:"lastname"`
	Created      time.Time `json:"created"`
	Modified     time.Time `json:"modified"`
}

// CustomerDevKey struct holding the details of a customer Developer Key including its bcrypt hash.
type CustomerDevKey struct {
	UUID         string    `json:"uuid"`
	Key          string    `json:"key"`
	CustomerUUID string    `json:"customer_uuid"`
	Created      time.Time `json:"created"`
	Modified     time.Time `json:"modified"`
}

// Address contains address information for a Customer
type Address struct {
	AddrUUID    string    `json:"addr_uuid"`
	Typ         string    `json:"typ"`
	ContactName string    `json:"contact_name"`
	Addr1       string    `json:"addr1"`
	Addr2       *string   `json:"addr2,omitempty"`
	City        string    `json:"city"`
	County      *string   `json:"county,omitempty"`
	Postcode    string    `json:"postcode"`
	Country     string    `json:"country"`
	Created     time.Time `json:"created"`
	Modified    time.Time `json:"modified"`
}

type CustomerService interface {
	CreateCustomer(ctx context.Context, role, email, password, firstname, lastname string) (*Customer, error)
	GetCustomers(ctx context.Context, p *PaginationQuery) (*PaginationResultSet, error)
	GetCustomer(ctx context.Context, customerUUID string) (*Customer, error)
	CreateAddress(ctx context.Context, customerUUID, typ, contactName, addr1 string, addr2 *string, city string, county *string, postcode string, country string) (*Address, error)
	GetAddress(ctx context.Context, uuid string) (*Address, error)
	GetAddressOwner(ctx context.Context, addrUUID string) (*string, error)
	GetAddresses(ctx context.Context, customerUUID string) ([]*Address, error)
	DeleteAddress(ctx context.Context, addrUUID string) error
	ListCustomersDevKeys(ctx context.Context, uuid string) ([]*CustomerDevKey, error)
	GenerateCustomerDevKey(ctx context.Context, uuid string) (*CustomerDevKey, error)
	SignInWithDevKey(ctx context.Context, key string) (string, error)
}

type PaginationQuery struct {
	OrderBy, OrderDir string
	Limit             int
	StartAfter        string
}
type PaginationContext struct {
	Total     int    `json:"total"`
	FirstUUID string `json:"first_uuid"`
	LastUUID  string `json:"last_uuid"`
}
type PaginationResultSet struct {
	RContext PaginationContext
	RSet     interface{}
}

type AuthService interface {
	Authenticate(ctx context.Context, jwt string) (*auth.Token, error)
}

type CatalogAndProductService interface {
	GetCatalog(ctx context.Context) ([]*nestedset.NestedSetNode, error)
	GetCatalogProductAssocs(ctx context.Context) ([]*model.CatalogProductAssoc, error)
	UpdateCatalogProductAssocs(ctx context.Context, cpo []*model.CatalogProductAssoc) error
}

type ErrorService interface {
	IsNotExist(err error) bool
}
