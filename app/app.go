package app

import (
	"bitbucket.org/andyfusniakteam/ecom-api-go/service/firebase"
)

// Cart operation sentinel values.
const (
	// Admin
	OpCreateAdmin string = "OpCreateAdmin"
	OpListAdmins  string = "OpListAdmins"
	OpDeleteAdmin string = "OpDeleteAdmin"

	// Cart
	OpCreateCart     string = "OpCreateCart"
	OpAddItemToCart  string = "OpAddItemToCart"
	OpGetCartItems   string = "OpGetCartItems"
	OpUpdateCartItem string = "OpUpdateCartItem"
	OpDeleteCartItem string = "OpDeleteCartItem"
	OpEmptyCartItems string = "OpEmptyCartItems"

	// Customers
	OpCreateCustomer        string = "OpCreateCustomer"
	OpGetCustomer           string = "OpGetCustomer"
	OpListCustomers         string = "OpListCustomers"
	OpGetCustomersAddresses string = "OpGetCustomersAddresses"
	OpUpdateAddress         string = "OpUpdateAddress"
	OpCreateAddress         string = "OpCreateAddress"
	OpGetAddress            string = "OpGetAddress"
	OpDeleteAddress         string = "OpDeleteAddress"

	// Products
	OpCreateProduct string = "OpCreateProduct"
	OpGetProduct    string = "OpGetProduct"
	OpListProducts  string = "OpListProducts"
	OpProductExists string = "OpProductExists"
	OpUpdateProduct string = "OpUpdateProduct"
	OpDeleteProduct string = "OpDeleteProduct"

	// Pricing
	OpGetTierPricing    string = "OpGetTierPricing"
	OpListPricingBySKU  string = "OpListPricingBySKU"
	OpListPricingByTier string = "OpListPricingByTier"
	OpUpdateTierPricing string = "OpUpdateTierPricing"
	OpDeleteTierPricing string = "OpDeleteTierPricing"

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
