package app

import (
	"net/http"

	"github.com/go-chi/chi"
	log "github.com/sirupsen/logrus"
)

// DeleteCartItemHandler creates a handler to delete an item from the cart with the given cart UUID.
func (a *App) DeleteCartItemHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		contextLogger := log.WithContext(ctx)
		contextLogger.Info("App: DeleteCartItemHandler started")

		uuid := chi.URLParam(r, "uuid")
		sku := chi.URLParam(r, "sku")
		count, err := a.Service.DeleteCartItem(ctx, uuid, sku)
		if count == 0 {
			w.WriteHeader(http.StatusNotFound) // 404 Not Found
			return
		}
		if err != nil {
			contextLogger.Errorf("DeleteCartItem(ctx, %q, %q) failed: %v", uuid, sku, err)
			w.WriteHeader(http.StatusInternalServerError) // 500 Internal Server Error
			return
		}
		w.Header().Del("Content-Type")
		w.WriteHeader(http.StatusNoContent) // 204 No Content
	}
}
