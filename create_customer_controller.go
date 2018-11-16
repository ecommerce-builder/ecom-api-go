package app

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
)

// CreateCustomer controller
func (a *App) CreateCustomerController() http.HandlerFunc {
	type createCustomerRequestBody struct {
		Email     string `json:"email"`
		Password  string `json:"password"`
		Firstname string `json:"firstname"`
		Lastname  string `json:"lastname"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
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
		err := json.NewDecoder(r.Body).Decode(&o)
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
		defer r.Body.Close()

		customer, err := a.Service.CreateCustomer(o.Email, o.Password, o.Firstname, o.Lastname)
		if err != nil {
			fmt.Fprintf(os.Stderr, "CreateCustomerController: failed Service.CreateCustomer(%s, %s, %s, %s): %v\n", o.Email, "*****", o.Firstname, o.Lastname, err)
			w.WriteHeader(http.StatusInternalServerError) // 500
			json.NewEncoder(w).Encode(struct {
				Code    int    `json:"code"`
				Message string `json:"message"`
			}{
				400,
				err.Error(),
			})
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated) // 201 Created
		json.NewEncoder(w).Encode(*customer)
	}
}
