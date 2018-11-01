package services

import "bitbucket.org/andyfusniakteam/ecom-api-go/models"

// CreateCart generates a new random UUID to be used for subseqent cart calls
func CreateCart() string {
	return models.CreateCart()
}

// AddItemToCart adds a single item to a given cart
func AddItemToCart(cartUUID string, sku string, qty int) (*models.CartItem, error) {
	item, err := models.AddItemToCart("default", cartUUID, sku, qty)
	if err != nil {
		panic(err)
	}

	return item, nil
}

// GetCartItems get all items in the given cart
func GetCartItems(cartUUID string) ([]models.CartItem, error) {
	return models.GetCartItems(cartUUID)
}

// UpdateCartItem updates a single item's qty
func UpdateCartItem(cartUUID string, sku string, qty int) (*models.CartItem, error) {
	return models.UpdateCartItemByUUID(cartUUID, sku, qty)
}

// DeleteCartItem deletes a single cart item
func DeleteCartItem(cartUUID string, sku string) (count int64, err error) {
	return models.DeleteCartItem(cartUUID, sku)
}

// EmptyCartItems empties the cart of all items but not coupons
func EmptyCartItems(cartUUID string) (err error) {
	return models.EmptyCartItems(cartUUID)
}
