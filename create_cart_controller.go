package app

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
)

// CreateCartController handler
func (a *App) CreateCartController() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var cart struct {
			CartUUID string `json:"cart_uuid"`
		}

		uuid, err := a.Service.CreateCart()
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to create cart: %v", err)
		}

		cart.CartUUID = *uuid

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated) // 201 Created
		json.NewEncoder(w).Encode(cart)
	}
}
