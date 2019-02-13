package app

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
)

// CreateAdminHandler creates a new administrator
func (a *App) CreateAdminHandler() http.HandlerFunc {
	type createAdminRequestBody struct {
		Email     string `json:"email"`
		Password  string `json:"password"`
		Firstname string `json:"firstname"`
		Lastname  string `json:"lastname"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		o := createAdminRequestBody{}
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

		customer, err := a.Service.CreateCustomer(r.Context(), "admin", o.Email, o.Password, o.Firstname, o.Lastname)
		if err != nil {
			fmt.Fprintf(os.Stderr, "CreateAdminHandler: failed Service.CreateCustomer(ctx, %q, %s, %s, %s, %s): %v\n", "admin", o.Email, "*****", o.Firstname, o.Lastname, err)
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

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated) // 201 Created
		json.NewEncoder(w).Encode(*customer)
	}
}
