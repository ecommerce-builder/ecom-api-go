package app

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
)

// CreateCartHandler create a new shopping cart
func (a *App) CreateCartHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var cart struct {
			UUID string `json:"uuid"`
		}
		uuid, err := a.Service.CreateCart(r.Context())
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to create cart: %v", err)
			w.WriteHeader(http.StatusInternalServerError) // 500 Internal Server Error
			return
		}
		cart.UUID = *uuid
		w.WriteHeader(http.StatusCreated) // 201 Created
		json.NewEncoder(w).Encode(cart)
	}
}
