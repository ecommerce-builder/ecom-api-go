package models

import (
	"database/sql"
)

// Cart structure is used to interact with a single shopping cart
type Cart struct {
	db *sql.DB
}

// CartItem structure holds the details individual cart ite
type CartItem struct {
	sku       string
	qty       int
	unitPrice float64
}

// CreateCart creates a new shopping cart
func CreateCart(db *sql.DB) *Cart {
	return &Cart{
		db: db,
	}
}

func (c *Cart) AddItemToCart(CartItem) {
	// rows, err := db.Query("SELECT name FROM users WHERE age = $1", age)
}
