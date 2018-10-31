package controllers

import (
	"encoding/json"
	"net/http"

	"bitbucket.org/andyfusniakteam/ecom-api-go/services"
	"github.com/gorilla/mux"
)

// GetCustomer controller
func GetCustomer(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)

	c := services.GetCustomer(params["cid"])

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(c)
}
