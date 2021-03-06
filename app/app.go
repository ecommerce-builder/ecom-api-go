package app

import (
	"encoding/json"
	"net/http"
	"regexp"

	"bitbucket.org/andyfusniakteam/ecom-api-go/service/firebase"
)

// Inventory
const (
	// ErrCodeInventoryNotFound is sent when failing to retrive any inventory for both list
	// and get operations.
	ErrCodeInventoryNotFound string = "inventory/inventory-not-found"
)

// Shipping Tariffs
const (
	// ErrCodeShippingTariffCodeExists error
	ErrCodeShippingTariffCodeExists string = "shipping-tariffs/shipping-tariff-code-exists"

	// ErrCodeShippingTariffNotFound error
	ErrCodeShippingTariffNotFound string = "shipping-tariffs/shipping-tariff-not-found"
)

// Product Set Items
const (
	// ErrCodeProductSetNotFound error
	ErrCodeProductSetNotFound string = "product-sets/product-set-not-found"
)

// Price Lists
const (
	OpCreatePriceList string = "OpCreatePriceList"
	OpGetPriceList    string = "OpGetPriceList"
	OpListPriceLists  string = "OpListPricingPriceLists"
	OpUpdatePriceList string = "OpUpdatePriceList"
	OpDeletePriceList string = "OpDeletePriceList"

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
)

// Product to product association groups
const (
	OpCreateProductToProductAssocGroup string = "OpCreateProductToProductAssocGroup"
	OpGetProductToProductAssocGroup    string = "OpGetProductToProductAssocGroup"
	OpListProductToProductAssocGroups  string = "OpListProductToProductAssocGroups"
	OpDeleteProductToProductAssocGroup string = "OpDeleteProductToProductAssocGroup"

	// ErrCodePPAssocGroupNotFound error
	ErrCodePPAssocGroupNotFound string = "product-to-product-assocs-groups/assoc-group-not-found"

	// ErrCodePPAssocGroupExists error
	ErrCodePPAssocGroupExists string = "product-to-product-assocs-groups/assoc-group-exists"

	// ErrCodePPAssocGroupContainsAssocs error
	ErrPPAssocGroupContainsAssocs string = "product-to-product-assocs-groups/assoc-group-contains-assocs"
)

// Product to products associations
const (
	OpBatchUpdateProductToProductAssocs string = "OpBatchUpdateProductToProductAssocs"
	OpGetProductToProductAssoc          string = "OpGetProductToProductAssoc"
	OpListProductToProductAssocs        string = "OpListProductToProductAssocs"
	OpDeleteProductToProductAssoc       string = "OpDeleteProductToProductAssoc"

	// ErrCodePPAssocNotFound error
	ErrCodePPAssocNotFound string = "product-to-product-assocs/assoc-not-found"
)

// Offers
const (
	// ErrCodeOfferNotFound error
	ErrCodeOfferNotFound string = "offers/offer-not-found"

	// ErrCodeOfferExists error
	ErrCodeOfferExists string = "offers/offer-exists"
)

// Coupons
const (
	OpCreateCoupon string = "OpCreateCoupon"
	OpGetCoupon    string = "OpGetCoupon"
	OpListCoupons  string = "OpListCoupons"
	OpUpdateCoupon string = "OpUpdateCoupon"
	OpDeleteCoupon string = "OpDeleteCoupon"

	// ErrCodeCouponNotFound error
	ErrCodeCouponNotFound string = "coupons/coupon-not-found"

	// ErrCodeCouponExists error
	ErrCodeCouponExists string = "coupons/coupon-exists"

	// ErrCodeCouponInUse occurs when trying to delete a coupon that is in a cart.
	ErrCodeCouponInUse string = "coupons/coupon-in-use"

	// ErrCodeCouponExpired error
	ErrCodeCouponExpired string = "coupons/coupon-expired"

	// ErrCodeCouponVoid error
	ErrCodeCouponVoid string = "coupons/coupon-void"

	// ErrCodeCouponUsed error
	ErrCodeCouponUsed string = "coupons/coupon-used"
)

// User operations and error codes
const (
	OpCreateUser string = "OpCreateUser"
	OpGetUser    string = "OpGetUser"
	OpListUsers  string = "OpListUsers"
	OpDeleteUser string = "OpDeleteUser"

	// ErrCodeUserNotFound is sent if a user cannot be located
	ErrCodeUserNotFound string = "users/user-not-found"

	// ErrCodeCreateUserForbidden error
	ErrCodeCreateUserForbidden string = "users/create-admin-forbidden"

	// ErrCodeUserExists error
	ErrCodeUserExists string = "users/user-exists"

	// ErrCodeUserInUse error
	ErrCodeUserInUse string = "users/user-in-use"
)

// Addresses
const (
	OpGetUsersAddresses string = "OpGetUsersAddresses"
	OpUpdateAddress     string = "OpUpdateAddress"
	OpCreateAddress     string = "OpCreateAddress"
	OpGetAddress        string = "OpGetAddress"
	OpDeleteAddress     string = "OpDeleteAddress"

	// ErrCodeAddressNotFound error
	ErrCodeAddressNotFound string = "addresses/address-not-found"
)

// Categories
const (
	// ErrCodeCategoryNotFound is sent when referencing a category with an invalid uuid.
	ErrCodeCategoryNotFound string = "categories/category-not-found"

	// ErrCodeCategoryNotLeaf is sent when trying to link a product to a non-leaf category.
	ErrCodeCategoryNotLeaf string = "categories/category-not-leaf"

	// ErrCodeLeafCategoryNotFound returned when attempting to associate a product to a
	// leaf category and that categoryd.
	ErrCodeLeafCategoryNotFound string = "categories/leaf-category-not-found"

	// ErrCodeCategoriesInUse occurs when attempting to update the categories tree and
	// one or more categories are referenced in a promo rule.
	ErrCodeCategoriesInUse string = "categories/categories-in-use"
)

// Orders
const (
	OpPlaceOrder string = "OpPlaceOrder"
	OpGetOrder   string = "OpGetOrder"
	OpListOrders string = "OpListOrders"

	// ErrCodeOrderCartEmpty error
	ErrCodeOrderCartEmpty string = "orders/order-cart-empty"

	// ErrCodeOrderUserNotFound error
	ErrCodeOrderUserNotFound string = "orders/order-user-not-found"

	// ErrCodeOrderNotFound error
	ErrCodeOrderNotFound string = "orders/order-not-found"

	// ErrCodeOrderItemsNotFound error
	ErrCodeOrderItemsNotFound string = "orders/order-items-not-found"
)

// Products
const (
	OpCreateProduct string = "OpCreateProduct"
	OpUpdateProduct string = "OpUpdateProduct"
	OpGetProduct    string = "OpGetProduct"
	OpListProducts  string = "OpListProducts"
	OpDeleteProduct string = "OpDeleteProduct"

	// ErrCodeProductNotFound indicates the product with given SKU could not be found.
	ErrCodeProductNotFound string = "products/product-not-found"

	// ErrCodeProductHasNoPrices occurs when attempting to perform an operation
	// that requires a product's price.
	ErrCodeProductHasNoPrices string = "products/product-has-no-prices"

	// ErrCodeProductPathExists is returned when attempting to create or update a product
	// with a path that is already used by another product.
	ErrCodeProductPathExists string = "products/product-path-exists"

	// ErrCodeProductSKUExists is returned when attempting to create or update a product
	// with a SKU that is already used by another product.
	ErrCodeProductSKUExists string = "products/product-sku-exists"
)

// Carts
const (
	OpCreateCart        string = "OpCreateCart"
	OpAddProductToCart  string = "OpAddProductToCart"
	OpGetCartProducts   string = "OpGetCartProducts"
	OpUpdateCartProduct string = "OpUpdateCartProduct"
	OpDeleteCartProduct string = "OpDeleteCartProduct"
	OpEmptyCartProducts string = "OpEmptyCartProducts"

	// ErrCodeCartProductExists is sent when attempting to add a product to a cart
	// and that product is already in the cart.
	ErrCodeCartProductExists string = "carts/cart-product-exists"

	// ErrCodeCartNotFound is sent when attempting to do cart operation of a non existing
	// cart.
	ErrCodeCartNotFound string = "carts/cart-not-found"
)

// Carts Coupons
const (
	OpApplyCouponToCart     string = "OpApplyCouponToCart"
	OpGetCartCoupon         string = "OpGetCartCoupon"
	OpListCartCoupons       string = "OpListCartCoupons"
	OpUnapplyCouponFromCart string = "OpUnapplyCouponFromCart"

	// ErrCodeCartCouponExists error
	ErrCodeCartCouponExists string = "carts-coupons/cart-coupon-exists"

	// ErrCodeCartCouponNotFound error
	ErrCodeCartCouponNotFound string = "carts-coupons/cart-coupon-not-found"

	// ErrCodeCouponNotAtStartDate error
	ErrCodeCouponNotAtStartDate string = "carts/coupon-not-at-start-date"
)

// Promo Rules
const (
	OpCreatePromoRule string = "OpCreatePromoRule"
	OpGetPromoRule    string = "OpGetPromoRule"
	OpListPromoRules  string = "OpListPromoRules"
	OpDeletePromoRule string = "OpDeletePromoRule"

	// ErrCodePromoRuleExists error
	ErrCodePromoRuleExists string = "promo-rules/promo-rule-exists"

	// ErrCodePromoRuleNotFound error
	ErrCodePromoRuleNotFound string = "promo-rules/promo-rule-not-found"
)

// Webhooks
const (
	OpCreateWebhook string = "OpCreateWebhook"
	OpGetWebhook    string = "OpGetWebhook"
	OpListWebhooks  string = "OpListWebhooks"
	OpUpdateWebhook string = "OpUpdateWebhook"
	OpDeleteWebhook string = "OpDeleteWebhook"

	// ErrCodeWebhookNotFound error
	ErrCodeWebhookNotFound string = "webhooks/webhook-not-found"

	// ErrCodeWebhookExists error
	ErrCodeWebhookExists string = "webhooks/webhook-exists"

	// ErrCodeEventTypeNotFound error
	ErrCodeEventTypeNotFound string = "webhooks/event-type-not-found"

	// ErrCodeWebhookPostFailed error
	ErrCodeWebhookPostFailed string = "webhooks/http-post-failed"
)

const (
	// ErrCodeInternalServerError is sent as the error code for 500 Internal Server Errors.
	ErrCodeInternalServerError string = "internal-server-error"

	// ErrCodeBadRequest is sent as the error code for 400 Bad Request.
	ErrCodeBadRequest string = "bad-request"

	// ErrCodeNotImplemented to indicate a operation is not yet implemented in code.
	ErrCodeNotImplemented string = "not-implemented"

	// ErrCodeAssocsExist is sent when the consumer attempts to purge the catalog
	// before purging the catalog associations.
	ErrCodeAssocsExist string = "assocs/assocs-exists"

	// ErrCodeNoCatalog is sent when the consumer attempts to apply catalog associations
	// before a catalog has been applied.
	ErrCodeNoCatalog string = "catalog/no-catalog"

	// ErrCodeCartProductNotFound is sent when attempting to delete or update a
	// cart product that cannot be found.
	ErrCodeCartProductNotFound string = "carts/cart-product-not-found"

	// ErrMissingPathsLeafsProductIDs is sent when the consumer attempts to apply catalog
	// associations that contain references to paths that are non existence, paths that
	// are non leaf categories or product SKUs that do not exist.
	ErrMissingPathsLeafsProductIDs string = "assocs/missing-paths-leafs-product-ids"

	// ErrCodeProductCategoryExists indicates the product to category assocation has
	// already been made.
	ErrCodeProductCategoryExists string = "product-category/product-category-exists"

	// ErrCodeProductCategoryNotFound for missing uuid.
	ErrCodeProductCategoryNotFound string = "product-category/product-category-not-found"

	// ErrCodeImageNotFound is returned whilst attempting to get an image.
	ErrCodeImageNotFound string = "images/image-not-found"

	// ErrCodeIncludeQueryParamParseError occurs when the include query param is invalid.
	ErrCodeIncludeQueryParamParseError string = "query/include-query-param-invalid"

	// ErrCodeAuthenticationFailed occurs when the Authentication has failed.
	ErrCodeAuthenticationFailed string = "auth/authentication-failed"

	// ErrCodeDeveloperKeyNotFound occurs when attempting to delete a developer key.
	ErrCodeDeveloperKeyNotFound string = "developer-keys/developer-key-not-found"
)

// Cart operation sentinel values.
const (
	// Prices
	OpUpdateProductPrices string = "OpUpdateProductPrices"
	OpGetProductPrices    string = "OpGetProductPrices"

	OpGetTierPricing    string = "OpGetTierPricing"
	OpMapPricingByTier  string = "OpMapPricingByTier"
	OpDeleteTierPricing string = "OpDeleteTierPricing"

	// Inventory
	OpGetInventory         string = "OpGetInventory"
	OpListInventory        string = "OpListInventory"
	OpUpdateInventory      string = "OpUpdateInventory"
	OpBatchUpdateInventory string = "OpBatchUpdateInventory"

	// Product Set Items
	OpGetProductSetItems string = "OpGetProductSetItems"

	// Shipping Tariffs
	OpCreateShippingTariff string = "OpCreateShippingTariff"
	OpGetShippingTariff    string = "OpGetShippingTariff"
	OpListShippingTariffs  string = "OpListShippingTariffs"
	OpUpdateShippingTariff string = "OpUpdateShippingTariff"
	OpDeleteShippingTariff string = "OpDeleteShippingTariff"

	// Offers
	OpActivateOffer   string = "OpActivateOffer"
	OpGetOffer        string = "OpGetOffer"
	OpListOffers      string = "OpListOffers"
	OpDeactivateOffer string = "OpDeactivateOffer"

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
	OpAddProductCategoryRelations    string = "OpAddProductCategoryRelations"
	OpGetProductCategoryRelations    string = "OpGetProductCategoryRelations"
	OpUpdateProductCategoryRelations string = "OpUpdateProductCategoryRelations"
	OpDeleteProductCategoryRelations string = "OpDeleteProductCategoryRelations"

	// Categories Tree
	OpGetCategoriesTree    string = "OpGetCategoriesTree"
	OpUpdateCategoriesTree string = "OpUpdateCategoriesTree"

	// Categories
	OpGetCategories    string = "OpGetCategories"
	OpDeleteCategories string = "OpDeleteCategories"

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

func clientError(w http.ResponseWriter, statusCode int, code string, message string) {
	// 4xx (Client Error): The request contains bad syntax or cannot be fulfilled
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(struct {
		Status  int    `json:"status"`
		Code    string `json:"code"`
		Message string `json:"message"`
	}{
		statusCode,
		code,
		message,
	})
}

func serverError(w http.ResponseWriter, statusCode int, code string, message string) {
	// 5xx (Server Error): The server failed to fulfill an apparently valid request
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(struct {
		Status  int    `json:"status"`
		Code    string `json:"code"`
		Message string `json:"message"`
	}{
		statusCode,
		code,
		message,
	})
}

// IsValidUUID checks for a valid UUID v4.
func IsValidUUID(uuid string) bool {
	r := regexp.MustCompile("^[a-fA-F0-9]{8}-[a-fA-F0-9]{4}-4[a-fA-F0-9]{3}-[8|9|aA|bB][a-fA-F0-9]{3}-[a-fA-F0-9]{12}$")
	return r.MatchString(uuid)
}
