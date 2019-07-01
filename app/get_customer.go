package app

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi"
	log "github.com/sirupsen/logrus"
)

// GetCustomerHandler returns an http.HandlerFunc that calls the service API
// to retrieve a list of Customers.
func (a *App) GetCustomerHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		contextLogger := log.WithContext(ctx)
		contextLogger.Info("App: GetCustomerHandler called")

		uuid := chi.URLParam(r, "uuid")
		customer, err := a.Service.GetCustomer(ctx, uuid)
		if err != nil {
			contextLogger.Errorf("service GetCustomer(%s) error: %+v", uuid, err)
			w.WriteHeader(http.StatusInternalServerError) // 500 Internal Server Error
			return
		}
		w.WriteHeader(http.StatusOK) // 200 OK
		json.NewEncoder(w).Encode(*customer)
	}
}
