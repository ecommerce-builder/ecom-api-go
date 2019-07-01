package app

import (
	"net/http"

	"github.com/go-chi/chi"
	log "github.com/sirupsen/logrus"
)

// DeleteCustomerDevKeyHandler deletes a customer Developer Key.
func (a *App) DeleteCustomerDevKeyHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		contextLogger := log.WithContext(ctx)
		contextLogger.Info("App: DeleteCustomerDevKeyHandler started")

		ctid := chi.URLParam(r, "ccuid")
		sku := chi.URLParam(r, "sku")
		count, _ := a.Service.DeleteCartItem(ctx, ctid, sku)
		if count == 0 {
			w.WriteHeader(http.StatusNotFound) // 404 Not Found
			return
		}
		w.Header().Del("Content-Type")
		w.WriteHeader(http.StatusNoContent) // 204 No Content
	}
}
