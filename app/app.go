package app

import (
	"bitbucket.org/andyfusniakteam/ecom-api-go/service/firebase"
)

// Cart operation sentinel values.
const (
	// Admin
	OpCreateAdmin string = "CreateAdmin"
	OpListAdmins  string = "ListAdmins"
	OpDeleteAdmin string = "DeleteAdmin"

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
	OpListProducts  string = "ListProducts"
	OpProductExists string = "ProductExists"
	OpUpdateProduct string = "UpdateProduct"
	OpDeleteProduct string = "DeleteProduct"

	// Developer Keys
	OpGenerateCustomerDevKey string = "GenerateCustomerDevKey"
	OpListCustomersDevKeys   string = "ListCustomersDevKeys"
	OpDeleteCustomerDevKey   string = "DeleteCustomerDevKey"
	OpSignInWithDevKey       string = "SignInWithDevKey"

	// Catalog and assocations
	OpUpdateCatalog       string = "UpdateCatalog"
	OpGetCatalog          string = "GetCatalog"
	OpPurgeCatalog        string = "PurgeCatalog"
	OpGetCatalogAssocs    string = "GetCatalogAssocs"
	OpUpdateCatalogAssocs string = "UpdateCatalogAssocs"
	OpPurgeCatalogAssocs  string = "PurgeCatalogAssocs"

	// System
	OpSystemInfo string = "SystemInfo"

	// Roles
	RoleSuperUser string = "root"
	RoleAdmin     string = "admin"
	RoleCustomer  string = "customer"
	RoleShopper   string = "anon"
)

// App defines the API application
type App struct {
	Service *firebase.Service
}
