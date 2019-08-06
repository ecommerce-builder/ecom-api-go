package app

import (
	"encoding/json"
	"net/http"

	service "bitbucket.org/andyfusniakteam/ecom-api-go/service/firebase"
	"github.com/go-chi/chi"
	log "github.com/sirupsen/logrus"
)

// ListCustomersDevKeysHandler get a list of addresses
func (a *App) ListCustomersDevKeysHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		contextLogger := log.WithContext(ctx)
		contextLogger.Info("App: ListCustomersDevKeysHandler started")

		customerID := chi.URLParam(r, "customer_id")
		apiKeys, err := a.Service.ListCustomersDevKeys(ctx, customerID)
		if err != nil {
			if err == service.ErrCustomerNotFound {
				w.WriteHeader(http.StatusNotFound) // 404 Not Found
				return
			}
			contextLogger.Errorf("service ListCustomersDevKeys(ctx, customerID=%q) error: %v", customerID, err)
			w.WriteHeader(http.StatusInternalServerError) // 500 Internal Server Error
			return
		}
		w.WriteHeader(http.StatusOK) // 200 OK
		json.NewEncoder(w).Encode(apiKeys)
	}
}
