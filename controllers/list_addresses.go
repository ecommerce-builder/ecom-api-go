package controllers

import (
	"encoding/json"
	"net/http"

	"bitbucket.org/andyfusniakteam/ecom-api-go/services"
	"github.com/gorilla/mux"
)

// ListAddresses handler
func ListAddresses(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)

	addresses, _ := services.GetAddresses(params["cid"])

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(addresses)
}
