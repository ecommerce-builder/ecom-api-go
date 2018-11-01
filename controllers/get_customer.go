package controllers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"bitbucket.org/andyfusniakteam/ecom-api-go/services"
	"github.com/gorilla/mux"
)

// GetCustomer controller
func GetCustomer(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)

	fmt.Println(r.Host)
	fmt.Println(r.URL)

	customer := services.GetCustomer(params["cid"])

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK) // 200 OK
	json.NewEncoder(w).Encode(customer)
}
