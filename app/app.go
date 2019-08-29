package app

import (
	"bitbucket.org/andyfusniakteam/ecom-api-go/service/firebase"
)

const (
	// ErrCodeInternalServerError is sent as the error code for 500 Internal Server Errors.
	ErrCodeInternalServerError string = "internal-server-error"

	// ErrCodeBadRequest is sent as the error code for 400 Bad Request.
	ErrCodeBadRequest string = "bad-request"

	// ErrCodeAssocsExist is sent when the consumer attempts to purge the catalog
	// before purging the catalog associations.
	ErrCodeAssocsExist string = "assocs/assocs-exists"

	// ErrCodeNoCatalog is sent when the consumer attempts to apply catalog associations
	// before a catalog has been applied.
	ErrCodeNoCatalog string = "catalog/no-catalog"

	// ErrCodeCategoryNotFound is sent when referencing a category with an invalid uuid.
	ErrCodeCategoryNotFound string = "category/category-not-found"

	// ErrCodeCategoryNotLeaf is sent when trying to link a product to a non-leaf category.
	ErrCodeCategoryNotLeaf string = "category/category-not-leaf"

	// ErrCodeLeafCategoryNotFound returned when attempting to associate a product to a
	// leaf category and that categoryd.
	ErrCodeLeafCategoryNotFound string = "categories/leaf-category-not-found"

	// ErrCodeUserNotFound is sent if a user cannot be located
	ErrCodeUserNotFound string = "users/not-found"

	// ErrCodeCartProductExists is sent when attempting to add a product to a cart
	// and that product is already in the cart.
	ErrCodeCartProductExists string = "carts/cart-product-exists"

	// ErrCodeCartProductNotFound is sent when attempting to delete or update a
	// cart product that cannot be found.
	ErrCodeCartProductNotFound string = "carts/cart-product-not-found"

	// ErrMissingPathsLeafsProductIDs is sent when the consumer attempts to apply catalog
	// associations that contain references to paths that are non existence, paths that
	// are non leaf categories or product SKUs that do not exist.
	ErrMissingPathsLeafsProductIDs string = "assocs/missing-paths-leafs-product-ids"

	// ErrCodeProductNotFound indicates the product with given SKU could not be found.
	ErrCodeProductNotFound string = "products/product-not-found"

	// ErrCodeProductCategoryExists indicates the product to category assocation has
	// already been made.
	ErrCodeProductCategoryExists string = "product-category/product-category-exists"

	// ErrCodeProductCategoryNotFound for missing uuid.
	ErrCodeProductCategoryNotFound string = "product-category/product-category-not-found"

	// ErrCodeProductPathExists is returned when attempting to create or update a product
	// with a path that is already used by another product.
	ErrCodeProductPathExists string = "products/product-path-exists"

	// ErrCodeProductSKUExists is returned when attempting to create or update a product
	// with a SKU that is already used by another product.
	ErrCodeProductSKUExists string = "products/product-sku-exists"

	// ErrCodeImageNotFound is returned whilst attempting to get an image.
	ErrCodeImageNotFound string = "images/image-not-found"

	// ErrCodePriceListNotFound is returned when attempting to reference a price
	// list that does not exist
	ErrCodePriceListNotFound string = "price-lists/price-list-not-found"

	// ErrCodePriceListCodeExists is returned when attempting to add a new price list
	// with a price list code that is already in use.
	ErrCodePriceListCodeExists string = "price-lists/price-list-code-exists"

	// ErrCodePriceListInUse is returned when attempting to delete a price list
	// that is already being referenced by prices.
	ErrCodePriceListInUse string = "price-lists/price-list-in-use"

	// ErrCodePriceListForbiddenPriceList for insufficient permissions to access price list.
	// determined by looking at the price list inside the customer object.
	ErrCodePriceListForbiddenPriceList string = "price-lists/forbidden-access-price-list"

	// ErrCodeIncludeQueryParamParseError occurs when the include query param is invalid.
	ErrCodeIncludeQueryParamParseError string = "query/include-query-param-invalid"

	// ErrCodeAuthenticationFailed occurs when the Authentication has failed.
	ErrCodeAuthenticationFailed string = "auth/authentication-failed"
)

// Cart operation sentinel values.
const (
	// Admin
	OpCreateAdmin string = "OpCreateAdmin"
	OpListAdmins  string = "OpListAdmins"
	OpDeleteAdmin string = "OpDeleteAdmin"

	// Cart
	OpCreateCart        string = "OpCreateCart"
	OpAddProductToCart  string = "OpAddProductToCart"
	OpGetCartProducts   string = "OpGetCartProducts"
	OpUpdateCartProduct string = "OpUpdateCartProduct"
	OpDeleteCartProduct string = "OpDeleteCartProduct"
	OpEmptyCartProducts string = "OpEmptyCartProducts"

	// Orders
	OpPlaceOrder string = "OpPlaceOrder"

	// Users
	OpCreateUser        string = "OpCreateUser"
	OpGetUser           string = "OpGetUser"
	OpListUsers         string = "OpListUsers"
	OpGetUsersAddresses string = "OpGetUsersAddresses"
	OpUpdateAddress     string = "OpUpdateAddress"
	OpCreateAddress     string = "OpCreateAddress"
	OpGetAddress        string = "OpGetAddress"
	OpDeleteAddress     string = "OpDeleteAddress"

	// Products
	OpCreateProduct string = "OpCreateProduct"
	OpUpdateProduct string = "OpUpdateProduct"
	OpGetProduct    string = "OpGetProduct"
	OpListProducts  string = "OpListProducts"
	OpDeleteProduct string = "OpDeleteProduct"

	// Price Lists
	OpCreatePriceList string = "OpCreatePriceList"
	OpGetPriceList    string = "OpGetPriceList"
	OpListPriceLists  string = "OpListPricingPriceLists"
	OpUpdatePriceList string = "OpUpdatePriceList"
	OpDeletePriceList string = "OpDeletePriceList"

	// Prices
	OpUpdateProductPrices string = "OpUpdateProductPrices"
	OpGetProductPrices    string = "OpGetProductPrices"

	OpGetTierPricing    string = "OpGetTierPricing"
	OpMapPricingByTier  string = "OpMapPricingByTier"
	OpDeleteTierPricing string = "OpDeleteTierPricing"

	// Promotion Rules
	OpCreatePromoRule string = "OpCreatePromoRule"
	OpGetPromoRule    string = "OpGetPromoRule"
	OpListPromoRules  string = "OpListPromoRules"
	OpDeletePromoRule string = "OpDeletePromoRule"

	// Image
	OpAddImage               string = "OpAddImage"
	OpGetImage               string = "OpGetImage"
	OpListProductImages      string = "OpListProductImages"
	OpDeleteImage            string = "OpDeleteImage"
	OpDeleteAllProductImages string = "OpDeleteAllProductImages"

	// Developer Keys
	OpGenerateUserDevKey string = "OpGenerateUserDevKey"
	OpListUsersDevKeys   string = "OpListUsersDevKeys"
	OpDeleteUserDevKey   string = "OpDeleteUserDevKey"
	OpSignInWithDevKey   string = "OpSignInWithDevKey"

	// Category and assocations
	OpAddProductCategoryRel    string = "OpAddProductCategoryRel"
	OpGetProductCategoryRel    string = "OpGetProductCategoryRel"
	OpDeleteProductCategoryRel string = "OpDeleteProductCategoryRel"

	OpGetProductCategoryRels    string = "OpGetProductCategoryRels"
	OpUpdateProductCategoryRels string = "OpUpdateProductCategoryRels"
	OpDeleteProductCategoryRels string = "OpPurgeProductsCategories"

	// Categories Tree
	OpGetCategoriesTree    string = "OpGetCategoriesTree"
	OpUpdateCategoriesTree string = "OpUpdateCategoriesTree"

	// Categories
	OpGetCategories    string = "OpGetCategories"
	OpDeleteCategories string = "OpDeleteCategories"

	OpGetProductCategoryRelations    string = "OpGetProductCategoryRelations"
	OpUpdateProductCategoryRelations string = "OpUpdateProductCategoryRelations"
	OpDeleteCategoryRelations        string = "OpDeleteCategoryRelations"

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
