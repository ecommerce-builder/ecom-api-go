package controllers

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"io/ioutil"
	"net/http"

	"bitbucket.org/andyfusniakteam/ecom-api-go/services"
	"cloud.google.com/go/logging"
)

// DB is the database handle
var DB *sql.DB

type createCustomerRequestBody struct {
	Email     string `json:"email"`
	Password  string `json:"password"`
	Firstname string `json:"firstname"`
	Lastname  string `json:"lastname"`
}

// CreateCustomer controller
func CreateCustomer(w http.ResponseWriter, r *http.Request) {
	b, _ := ioutil.ReadAll(r.Body)

	Lg.Log(logging.Entry{Payload: struct {
		Method string
		URL    string
		Body   string
		Header http.Header
	}{r.Method, r.URL.String(), string(b), r.Header}, Severity: logging.Debug})

	if r.Body == nil {
		w.WriteHeader(http.StatusBadRequest) // 400 Bad Request
		json.NewEncoder(w).Encode(struct {
			Code    int    `json:"code"`
			Message string `json:"message"`
		}{
			400,
			"Please send a request body",
		})
		return
	}

	o := createCustomerRequestBody{}
	err := json.NewDecoder(bytes.NewBuffer(b)).Decode(&o)
	defer r.Body.Close()
	if err != nil {
		w.WriteHeader(http.StatusBadRequest) // 400 Bad Request
		json.NewEncoder(w).Encode(struct {
			Code    int    `json:"code"`
			Message string `json:"message"`
		}{
			400,
			err.Error(),
		})
		return
	}

	customer, _ := services.CreateCustomer(o.Email, o.Password, o.Firstname, o.Lastname)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated) // 201 Created
	json.NewEncoder(w).Encode(customer)
}
