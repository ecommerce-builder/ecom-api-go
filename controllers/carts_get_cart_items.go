package controllers

import (
	"encoding/json"
	"net/http"

	"bitbucket.org/andyfusniakteam/ecom-api-go/services"
	"github.com/gorilla/mux"
)

// GetCartItems handler
func GetCartItems(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)

	cartItems, err := services.GetCartItems(params["ctid"])
	if err != nil {
		panic(err)
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK) // 200 OK
	json.NewEncoder(w).Encode(cartItems)
}
