package app

import (
	"encoding/json"
	"net/http"

	service "bitbucket.org/andyfusniakteam/ecom-api-go/service/firebase"
	"github.com/go-chi/chi"
	log "github.com/sirupsen/logrus"
)

// GenerateCustomerDevKeyHandler creates a new API Key for a given customer
func (a *App) GenerateCustomerDevKeyHandler() http.HandlerFunc {
	type customerDevKeyResponseBody struct {
		Object string `json:"object"`
		*service.CustomerDevKey
	}

	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		contextLogger := log.WithContext(ctx)
		contextLogger.Info("App: GenerateCustomerDevKeyHandler started")

		customerID := chi.URLParam(r, "customer_id")
		cdk, err := a.Service.GenerateCustomerDevKey(ctx, customerID)
		if err != nil {
			if err == service.ErrCustomerNotFound {
				w.WriteHeader(http.StatusNotFound) // 404 Not Found
				return
			}
			contextLogger.Errorf("service GenerateCustomerAPIKey(ctx, customerID=%q) error: %v", customerID, err)
			w.WriteHeader(http.StatusInternalServerError) // 500 Internal Server Error
			return
		}
		res := customerDevKeyResponseBody{
			Object:         "developer_key",
			CustomerDevKey: cdk,
		}
		w.WriteHeader(http.StatusCreated) // 201 Created
		json.NewEncoder(w).Encode(res)
	}
}
