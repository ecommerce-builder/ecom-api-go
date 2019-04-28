package app

import (
	"bitbucket.org/andyfusniakteam/ecom-api-go/service/firebase"
)

// Cart operation sentinel values.
const (
	// Admin
	OpCreateAdmin string = "CreateAdmin"

	// Cart
	OpCreateCart     string = "CreateCart"
	OpAddItemToCart  string = "AddItemToCart"
	OpGetCartItems   string = "GetCartItems"
	OpUpdateCartItem string = "UpdateCartItem"
	OpDeleteCartItem string = "DeleteCartItem"
	OpEmptyCartItems string = "EmptyCartItems"

	// Customers
	OpCreateCustomer        string = "CreateCustomer"
	OpGetCustomer           string = "GetCustomer"
	OpListCustomers         string = "ListCustomers"
	OpGetCustomersAddresses string = "GetCustomersAddresses"
	OpUpdateAddress         string = "UpdateAddress"
	OpCreateAddress         string = "CreateAddress"
	OpGetAddress            string = "GetAddress"
	OpDeleteAddress         string = "DeleteAddress"

	// Products
	OpCreateProduct string = "CreateProduct"
	OpGetProduct    string = "GetProduct"
	OpProductExists string = "ProductExists"
	OpUpdateProduct string = "UpdateProduct"
	OpDeleteProduct string = "DeleteProduct"

	// Developer Keys
	OpGenerateCustomerDevKey string = "GenerateCustomerDevKey"
	OpListCustomersDevKeys   string = "ListCustomersDevKeys"
	OpDeleteCustomerDevKey   string = "DeleteCustomerDevKey"
	OpSignInWithDevKey       string = "SignInWithDevKey"

	// Catalog
	OpGetCatalog string = "GetCatalog"

	// System
	OpSystemInfo string = "SystemInfo"

	// Roles
	RoleSuperUser string = "root"
	RoleAdmin     string = "admin"
	RoleCustomer  string = "customer"
	RoleShopper   string = "anon"
)

type App struct {
	Service *firebase.Service
}
