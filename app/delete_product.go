package app

import (
	"net/http"

	"github.com/go-chi/chi"
	log "github.com/sirupsen/logrus"
)

// DeleteProductHandler create a handler to delete a product resource.
func (a *App) DeleteProductHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		contextLogger := log.WithContext(ctx)
		contextLogger.Info("App: DeleteProductHandler started")

		sku := chi.URLParam(r, "sku")
		if err := a.Service.DeleteProduct(ctx, sku); err != nil {
			contextLogger.Errorf("a.Service.DeleteProduct(ctx, sku=%q) failed: %v", sku, err)
			w.WriteHeader(http.StatusInternalServerError) // 500 Internal Server Error
			return
		}
		w.Header().Del("Content-Type")
		w.WriteHeader(http.StatusNoContent) // 204 No Content
	}
}
