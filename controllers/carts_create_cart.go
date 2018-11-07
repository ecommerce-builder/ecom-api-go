package controllers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"bitbucket.org/andyfusniakteam/ecom-api-go/services"
	"cloud.google.com/go/logging"
)

// Lg provides a global logger
var Lg *logging.Logger

// CreateCart handler
func CreateCart(w http.ResponseWriter, r *http.Request) {
	Lg.Log(logging.Entry{Payload: r.Header, Severity: logging.Debug})
	var cart struct {
		CartUUID string `json:"cart_uuid"`
	}

	uuid := services.CreateCart()
	Lg.Log(logging.Entry{Payload: fmt.Sprintf("services.CreateCart() returned %s", uuid), Severity: logging.Debug})

	cart.CartUUID = uuid

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated) // 201 Created
	json.NewEncoder(w).Encode(cart)
}
