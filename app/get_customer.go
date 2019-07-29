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

		id := chi.URLParam(r, "id")
		customer, err := a.Service.GetCustomer(ctx, id)
		if err != nil {
			contextLogger.Errorf("service GetCustomer(id=%s) error: %+v", id, err)
			w.WriteHeader(http.StatusInternalServerError) // 500 Internal Server Error
			return
		}
		w.WriteHeader(http.StatusOK) // 200 OK
		json.NewEncoder(w).Encode(*customer)
	}
}
