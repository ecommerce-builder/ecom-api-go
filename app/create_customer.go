package app

import (
	"encoding/json"
	"net/http"

	service "bitbucket.org/andyfusniakteam/ecom-api-go/service/firebase"
	log "github.com/sirupsen/logrus"
)

type customerResponseBody struct {
	Object string `json:"object"`
	*service.Customer
}

// CreateCustomerHandler creates a new customer record
func (a *App) CreateCustomerHandler() http.HandlerFunc {
	type createCustomerRequestBody struct {
		Email     string `json:"email"`
		Password  string `json:"password"`
		Firstname string `json:"firstname"`
		Lastname  string `json:"lastname"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		contextLogger := log.WithContext(ctx)
		contextLogger.Info("App: CreateCustomerHandler called")

		if r.Body == nil {
			w.WriteHeader(http.StatusBadRequest) // 400 Bad Request
			json.NewEncoder(w).Encode(struct {
				Status  int    `json:"status"`
				Code    string `json:"code"`
				Message string `json:"message"`
			}{
				http.StatusBadRequest,
				ErrCodeBadRequest,
				"missing request body",
			})
			return
		}
		o := createCustomerRequestBody{}
		err := json.NewDecoder(r.Body).Decode(&o)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest) // 400 Bad Request
			json.NewEncoder(w).Encode(struct {
				Status  int    `json:"status"`
				Code    string `json:"code"`
				Message string `json:"message"`
			}{
				http.StatusBadRequest,
				ErrCodeBadRequest,
				err.Error(),
			})
			return
		}
		defer r.Body.Close()

		customer, err := a.Service.CreateCustomer(ctx, "customer", o.Email, o.Password, o.Firstname, o.Lastname)
		if err != nil {
			contextLogger.Errorf("CreateCustomerHandler: failed Service.CreateCustomer(ctx, %q, %s, %s, %s, %s) with error: %v", "customer", o.Email, "*****", o.Firstname, o.Lastname, err)
			w.WriteHeader(http.StatusInternalServerError) // 500
			json.NewEncoder(w).Encode(struct {
				Status  int    `json:"status"`
				Code    string `json:"code"`
				Message string `json:"message"`
			}{
				http.StatusInternalServerError,
				ErrCodeInternalServerError,
				err.Error(),
			})
			return
		}

		res := customerResponseBody{
			Object:   "customer",
			Customer: customer,
		}
		w.WriteHeader(http.StatusCreated) // 201 Created
		json.NewEncoder(w).Encode(res)
	}
}
