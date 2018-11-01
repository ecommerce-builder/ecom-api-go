package controllers

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"bitbucket.org/andyfusniakteam/ecom-api-go/services"
)

// DB is the database handle
var DB *sql.DB

type createCustomerRequestBody struct {
	Firstname string `json:"firstname"`
	Lastname  string `json:"lastname"`
}

// CreateCustomer controller
func CreateCustomer(w http.ResponseWriter, r *http.Request) {
	if r.Body == nil {
		http.Error(w, "Please send a request body", 400)
		return
	}

	o := createCustomerRequestBody{}
	err := json.NewDecoder(r.Body).Decode(&o)
	if err != nil {
		http.Error(w, err.Error(), 400)
		return
	}

	customer, _ := services.CreateCustomer(o.Firstname, o.Lastname)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated) // 201 Created
	json.NewEncoder(w).Encode(customer)
}
