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

		id := chi.URLParam(r, "id")
		apiKeys, err := a.Service.ListCustomersDevKeys(ctx, id)
		if err != nil {
			contextLogger.Errorf("service ListCustomersDevKeys(ctx, %s) error: %v", id, err)
			w.WriteHeader(http.StatusInternalServerError) // 500 Internal Server Error
			return
		}
		w.WriteHeader(http.StatusOK) // 200 OK
		json.NewEncoder(w).Encode(apiKeys)
	}
}
