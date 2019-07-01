package app

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi"
	log "github.com/sirupsen/logrus"
)

// ListAddressesHandler get a list of addresses
func (a *App) ListAddressesHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		contextLogger := log.WithContext(ctx)
		contextLogger.Info("App: ListAddressesHandler started")

		uuid := chi.URLParam(r, "uuid")
		addresses, _ := a.Service.GetAddresses(ctx, uuid)
		w.WriteHeader(http.StatusOK) // 200 OK
		json.NewEncoder(w).Encode(addresses)
	}
}
