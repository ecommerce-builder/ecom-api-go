package app

import (
	"encoding/json"
	"net/http"

	service "bitbucket.org/andyfusniakteam/ecom-api-go/service/firebase"
	log "github.com/sirupsen/logrus"
)

// ListAddressesHandler get a list of addresses
func (a *App) ListAddressesHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		contextLogger := log.WithContext(ctx)
		contextLogger.Info("app: ListAddressesHandler started")

		userID := r.URL.Query().Get("user_id")
		if userID == "" {
			// TODO: check admin privileges and return the whole list
			serverError(w, http.StatusNotImplemented, ErrCodeNotImplemented, "not implemented")
			return
		}

		addresses, err := a.Service.GetAddresses(ctx, userID)
		if err != nil {
			if err == service.ErrAddressNotFound {
				clientError(w, http.StatusNotFound, ErrCodeAddressNotFound, "address not found")
				return
			}
			contextLogger.Errorf("a.Service.GetAddresses(ctx, customerID=%q) error: %v", userID, err)
			w.WriteHeader(http.StatusInternalServerError) // 500 Internal Server Error
			return
		}
		w.WriteHeader(http.StatusOK) // 200 OK
		json.NewEncoder(w).Encode(addresses)
	}
}
