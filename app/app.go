package app

import (
	"bitbucket.org/andyfusniakteam/ecom-api-go/service/firebase"
)

const (
	// ErrCodeInternalServerError is sent as the error code for 500 Internal Server Errors.
	ErrCodeInternalServerError string = "internal-server-error"

	// ErrCodeBadRequest is sent as the error code for 400 Bad Request.
	ErrCodeBadRequest string = "bad-request"

	// ErrCodeAssocsAlreadyExist is sent when the consumer attempts to purge the catalog
	// before purging the catalog associations.
	ErrCodeAssocsAlreadyExist string = "assocs/assocs-already-exist"

	// ErrCodeNoCatalog is sent when the consumer attempts to apply catalog associations
	// before a catalog has been applied.
	ErrCodeNoCatalog string = "catalog/no-catalog"

	// ErrCodeCartAlreadyExists is sent when attempting to add a cart item to a cart
	// and that cart item is already in the cart.
	ErrCodeCartAlreadyExists string = "cart/cart-item-already-exists"

	// ErrMissingPathsLeafsSKUs is sent when the consumer attempts to apply catalog
	// associations that contain references to paths that are non existence, paths that
	// are non leaf categories or product SKUs that do not exist.
	ErrMissingPathsLeafsSKUs string = "assocs/missing-paths-leafs-skus"

	// ErrCodeProductSKUNotFound indicates the product with given SKU could not be found.
	ErrCodeProductSKUNotFound string = "products/product-sku-not-found"

	// ErrCodeDuplicateImagePath return value is sent when a consumer attempts
	// add a new product with duplicate image paths.
	ErrCodeDuplicateImagePath string = "products/duplicate-image-paths"
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

	// Orders
	OpPlaceOrder string = "OpPlaceOrder"

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
	OpReplaceProduct string = "OpReplaceProduct"
	OpGetProduct     string = "OpGetProduct"
	OpListProducts   string = "OpListProducts"
	OpProductExists  string = "OpProductExists"
	OpDeleteProduct  string = "OpDeleteProduct"

	// Pricing Tiers
	OpCreateTier string = "OpCreateTier"
	OpGetTier    string = "OpGetTier"
	OpListTiers  string = "OpListTiers"
	OpUpdateTier string = "OpUpdateTier"
	OpDeleteTier string = "OpDeleteTier"

	// Pricing
	OpGetTierPricing    string = "OpGetTierPricing"
	OpMapPricingBySKU   string = "OpMapPricingBySKU"
	OpMapPricingByTier  string = "OpMapPricingByTier"
	OpUpdateTierPricing string = "OpUpdateTierPricing"
	OpDeleteTierPricing string = "OpDeleteTierPricing"

	// Image
	OpAddImage               string = "OpAddImage"
	OpGetImage               string = "OpGetImage"
	OpListProductImages      string = "OpListProductImages"
	OpDeleteImage            string = "OpDeleteImage"
	OpDeleteAllProductImages string = "OppDeleteAllProductImages"

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
