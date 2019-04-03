package app

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
)

// CreateCustomerHandler creates a new customer record
func (a *App) CreateCustomerHandler() http.HandlerFunc {
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

		customer, err := a.Service.CreateCustomer(r.Context(), "customer", o.Email, o.Password, o.Firstname, o.Lastname)
		if err != nil {
			fmt.Fprintf(os.Stderr, "CreateCustomerHandler: failed Service.CreateCustomer(ctx, %q, %s, %s, %s, %s): %v\n", "customer", o.Email, "*****", o.Firstname, o.Lastname, err)
			w.WriteHeader(http.StatusInternalServerError) // 500
			json.NewEncoder(w).Encode(struct {
				Code    int    `json:"code"`
				Message string `json:"message"`
			}{
				500,
				err.Error(),
			})
			return
		}
		w.WriteHeader(http.StatusCreated) // 201 Created
		json.NewEncoder(w).Encode(*customer)
	}
}
