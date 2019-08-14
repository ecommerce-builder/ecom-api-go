package app

import (
	"net/http"

	log "github.com/sirupsen/logrus"
)

// DeleteCustomerDevKeyHandler deletes a customer Developer Key.
func (a *App) DeleteCustomerDevKeyHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		contextLogger := log.WithContext(ctx)
		contextLogger.Info("App: DeleteCustomerDevKeyHandler started")

		w.WriteHeader(http.StatusNotImplemented) // 501 Not Implemented
		return
	}
}
