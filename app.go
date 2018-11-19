package app

import (
	"net/http"
	"time"

	"firebase.google.com/go/auth"
)

type Serverable interface {
	Controllerable
	EcomService
}

type EcomService interface {
	CartService
	CustomerService
	AuthService
}

type App struct {
	Service EcomService
}

type Controllerable interface {
	CreateCartController() http.HandlerFunc
	GetCartItemsController() http.HandlerFunc
	CreateCustomerController() http.HandlerFunc
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
	CreateCart() (*string, error)
	AddItemToCart(cartUUID, sku string, qty int) (*CartItem, error)
	GetCartItems(cartUUID string) ([]*CartItem, error)
	UpdateCartItem(cartUUID string, sku string, qty int) (*CartItem, error)
	DeleteCartItem(cartUUID string, sku string) (count int64, err error)
	EmptyCartItems(cartUUID string) (err error)
}

// Customer details
type Customer struct {
	CustomerUUID string    `json:"customer_uuid"`
	UID          string    `json:"uid"`
	Email        string    `json:"email"`
	Firstname    string    `json:"firstname"`
	Lastname     string    `json:"lastname"`
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
	CreateCustomer(email, password, firstname, lastname string) (*Customer, error)
	GetCustomer(customerUUID string) (*Customer, error)
	CreateAddress(customerUUID, typ, contactName, addr1 string, addr2 *string, city string, county *string, postcode string, country string) (*Address, error)
	GetAddress(UID string, addressUUID string) (*Address, error)
	GetAddresses(customerUUID string) ([]*Address, error)
	DeleteAddress(addrUUID string) error
}

type AuthService interface {
	Authenticate(jwt string) (*auth.Token, error)
}

func (c *CartItem) String() string {
	return "test"
}
