package app

import (
	"encoding/json"
	"net/http"

	log "github.com/sirupsen/logrus"
)

// ListAddressesHandler get a list of addresses
func (a *App) ListAddressesHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		contextLogger := log.WithContext(ctx)
		contextLogger.Info("App: ListAddressesHandler started")

		userID := r.URL.Query().Get("user_id")

		addresses, err := a.Service.GetAddresses(ctx, userID)
		if err != nil {
			contextLogger.Errorf("a.Service.GetAddresses(ctx, customerID=%q) error: %v", userID, err)
			w.WriteHeader(http.StatusInternalServerError) // 500 Internal Server Error
			return
		}
		w.WriteHeader(http.StatusOK) // 200 OK
		json.NewEncoder(w).Encode(addresses)
	}
}
