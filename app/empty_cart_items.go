package app

import (
	"net/http"

	"github.com/go-chi/chi"
	log "github.com/sirupsen/logrus"
)

// EmptyCartItemsHandler empties the cart of all items. This does not remove coupons.
func (a *App) EmptyCartItemsHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		contextLogger := log.WithContext(ctx)
		contextLogger.Info("App: EmptyCartItemsHandler started")

		id := chi.URLParam(r, "id")
		if err := a.Service.EmptyCartItems(ctx, id); err != nil {
			contextLogger.Errorf("service EmptyCartItems(ctx, %s) error: %v", id, err)
			w.WriteHeader(http.StatusInternalServerError) // 500 Internal Server Error
			return
		}
		w.Header().Del("Content-Type")
		w.WriteHeader(http.StatusNoContent) // 204 No Content
	}
}
