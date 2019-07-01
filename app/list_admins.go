package app

import (
	"encoding/json"
	"net/http"

	log "github.com/sirupsen/logrus"
)

// ListAdminsHandler creates a handler that returns a list of administrators.
func (a *App) ListAdminsHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		contextLogger := log.WithContext(ctx)
		contextLogger.Info("App: ListAdminsHandler started")

		admins, err := a.Service.ListAdmins(ctx)
		if err != nil {
			contextLogger.Errorf("service: GetAdmins(ctx) error: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK) // 200 OK
		json.NewEncoder(w).Encode(admins)
	}
}
