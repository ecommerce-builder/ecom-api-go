package app

import (
	"net/http"

	service "bitbucket.org/andyfusniakteam/ecom-api-go/service/firebase"
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

		productID := chi.URLParam(r, "id")
		if err := a.Service.DeleteAllProductImages(ctx, productID); err != nil {
			if err == service.ErrProductNotFound {
				w.WriteHeader(http.StatusNotFound) // 404 Not Found
				return
			}
			contextLogger.Errorf("DeleteAllProductImages(ctx, productID=%q) failed: %v", productID, err)
			w.WriteHeader(http.StatusInternalServerError) // 500 Internal Server Error
			return
		}
		w.Header().Del("Content-Type")
		w.WriteHeader(http.StatusNoContent) // 204 No Content
	}
}
