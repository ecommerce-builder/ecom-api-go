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
	ErrCodeAssocsAlreadyExist string = "assocs/assocs-already-exists"

	// ErrCodeNoCatalog is sent when the consumer attempts to apply catalog associations
	// before a catalog has been applied.
	ErrCodeNoCatalog string = "catalog/no-catalog"

	// ErrCodeCustomerNotFound is sent if a customer cannot be located
	ErrCodeCustomerNotFound string = "customer/not-found"

	// ErrCodeCartAlreadyExists is sent when attempting to add a cart item to a cart
	// and that cart item is already in the cart.
	ErrCodeCartAlreadyExists string = "cart/cart-item-already-exists"

	// ErrCodeCartContainsNoItems is sent when attempting to empty a cart that is
	// already empty.
	ErrCodeCartContainsNoItems string = "cart/cart-contains-no-items"

	// ErrCodeCartItemNotFound is sent when attempting to delete or update a product
	// in a given cart. The cart and product are found but the product is not in
	// the cart items.
	ErrCodeCartItemNotFound string = "cart/cart-item-not-found"

	// ErrMissingPathsLeafsProductIDs is sent when the consumer attempts to apply catalog
	// associations that contain references to paths that are non existence, paths that
	// are non leaf categories or product SKUs that do not exist.
	ErrMissingPathsLeafsProductIDs string = "assocs/missing-paths-leafs-product-ids"

	// ErrCodeProductNotFound indicates the product with given SKU could not be found.
	ErrCodeProductNotFound string = "product/product-not-found"

	// ErrCodeDuplicateImagePath return value is sent when a consumer attempts
	// add a new product with duplicate image paths.
	ErrCodeDuplicateImagePath string = "product/duplicate-image-paths"

	// ErrCodePriceListNotFound is returned when attempting to reference a price
	// list that does not exist
	ErrCodePriceListNotFound string = "price-list/price-list-not-found"

	// ErrCodePriceListCodeTaken is returned when attempting to add a new price list
	// with a price list code that is already in use.
	ErrCodePriceListCodeTaken string = "price-list/price-list-code-taken"

	// ErrCodePriceListInUse is returned when attempting to delete a price list
	// that is already being referenced by prices.
	ErrCodePriceListInUse string = "price-list/price-list-in-use"
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
	OpCreateProduct string = "OpCreateProduct"
	OpUpdateProduct string = "OpUpdateProduct"
	OpGetProduct    string = "OpGetProduct"
	OpListProducts  string = "OpListProducts"
	OpProductExists string = "OpProductExists"
	OpDeleteProduct string = "OpDeleteProduct"

	// Price Lists
	OpCreatePriceList string = "OpCreatePriceList"
	OpGetPriceList    string = "OpGetPriceList"
	OpListPriceLists  string = "OpListPricingPriceLists"
	OpUpdatePriceList string = "OpUpdatePriceList"
	OpDeletePriceList string = "OpDeletePriceList"

	// Pricing
	OpGetTierPricing        string = "OpGetTierPricing"
	OpMapPricingByProductID string = "OpMapPricingByProductID"
	OpMapPricingByTier      string = "OpMapPricingByTier"
	OpUpdateTierPricing     string = "OpUpdateTierPricing"
	OpDeleteTierPricing     string = "OpDeleteTierPricing"

	// Image
	OpAddImage               string = "OpAddImage"
	OpGetImage               string = "OpGetImage"
	OpListProductImages      string = "OpListProductImages"
	OpDeleteImage            string = "OpDeleteImage"
	OpDeleteAllProductImages string = "OpDeleteAllProductImages"

	// Developer Keys
	OpGenerateCustomerDevKey string = "OpGenerateCustomerDevKey"
	OpListCustomersDevKeys   string = "OpListCustomersDevKeys"
	OpDeleteCustomerDevKey   string = "OpDeleteCustomerDevKey"
	OpSignInWithDevKey       string = "OpSignInWithDevKey"

	// Category and assocations
	OpUpdateCategories            string = "OpUpdateCategories"
	OpGetCategories               string = "OpGetCategories"
	OpPurgeCatalog                string = "OpPurgeCatalog"
	OpGetCategoryProductAssocs    string = "OpGetCategoryProductAssocs"
	OpUpdateCategoryProductAssocs string = "OpUpdateCategoryProductAssocs"
	OpPurgeCategoryAssocs         string = "OpPurgeCategoryAssocs"

	// Stripe
	OpStripeCheckout string = "OpStripeCheckout"
	OpStripeWebhook  string = "OpStripeWebhook"

	// System
	OpSystemInfo string = "OpSysInfo"

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
