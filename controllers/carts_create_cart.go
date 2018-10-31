package controllers

import (
	"encoding/json"
	"net/http"

	"bitbucket.org/andyfusniakteam/ecom-api-go/services"
)

// CreateCart handler
func CreateCart(w http.ResponseWriter, r *http.Request) {
	var cart struct {
		CartUUID string `json:"cart_uuid"`
	}

	uuid := services.CreateCart()
	cart.CartUUID = uuid

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated) // 201 Created
	json.NewEncoder(w).Encode(cart)
}
