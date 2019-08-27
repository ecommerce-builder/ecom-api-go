package app

import (
	"net/http"

	log "github.com/sirupsen/logrus"
)

// DeleteUserDevKeyHandler deletes a customer Developer Key.
func (a *App) DeleteUserDevKeyHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		contextLogger := log.WithContext(ctx)
		contextLogger.Info("App: DeleteUserDevKeyHandler started")

		w.WriteHeader(http.StatusNotImplemented) // 501 Not Implemented
		return
	}
}
