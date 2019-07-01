package app

import (
	"net/http"

	"github.com/go-chi/chi"
	log "github.com/sirupsen/logrus"
)

// DeleteAllProductImagesHandler create a handler that deletes all images for the
// product with the given SKU.
func (a *App) DeleteAllProductImagesHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		contextLogger := log.WithContext(ctx)
		contextLogger.Info("App: DeleteAllProductImagesHandler started")

		sku := chi.URLParam(r, "sku")
		exists, err := a.Service.ProductExists(ctx, sku)
		if err != nil {
			contextLogger.Errorf("ProductExists(ctx, %q) failed: %v", sku, err)
			w.WriteHeader(http.StatusInternalServerError) // 500 Internal Server Error
			return
		}
		if !exists {
			w.WriteHeader(http.StatusNotFound) // 404 Not Found
			return
		}
		if err := a.Service.DeleteAllProductImages(ctx, sku); err != nil {
			contextLogger.Errorf("DeleteAllProductImages(ctx, %q) failed: %v", sku, err)
			w.WriteHeader(http.StatusInternalServerError) // 500 Internal Server Error
			return
		}
		w.Header().Del("Content-Type")
		w.WriteHeader(http.StatusNoContent) // 204 No Content
	}
}
