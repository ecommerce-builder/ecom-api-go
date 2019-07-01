package app

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi"
	log "github.com/sirupsen/logrus"
)

// ListCustomersDevKeysHandler get a list of addresses
func (a *App) ListCustomersDevKeysHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		contextLogger := log.WithContext(ctx)
		contextLogger.Info("App: ListCustomersDevKeysHandler started")

		uuid := chi.URLParam(r, "uuid")
		apiKeys, err := a.Service.ListCustomersDevKeys(ctx, uuid)
		if err != nil {
			contextLogger.Errorf("service ListCustomersDevKeys(ctx, %s) error: %v", uuid, err)
			w.WriteHeader(http.StatusInternalServerError) // 500 Internal Server Error
			return
		}
		w.WriteHeader(http.StatusOK) // 200 OK
		json.NewEncoder(w).Encode(apiKeys)
	}
}
