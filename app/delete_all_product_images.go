package app

import (
	"net/http"

	service "bitbucket.org/andyfusniakteam/ecom-api-go/service/firebase"
	log "github.com/sirupsen/logrus"
)

// DeleteAllProductImagesHandler create a handler that deletes all images for the
// product with the given SKU.
func (a *App) DeleteAllProductImagesHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		contextLogger := log.WithContext(ctx)
		contextLogger.Info("app: DeleteAllProductImagesHandler started")

		productID := r.URL.Query().Get("product_id")
		if productID == "" {
			clientError(w, http.StatusBadRequest, ErrCodeBadRequest, "query parameter product_id must be set")
			return
		}

		if err := a.Service.DeleteAllProductImages(ctx, productID); err != nil {
			if err == service.ErrProductNotFound {
				clientError(w, http.StatusNotFound, ErrCodeProductNotFound, "product not found")
				return
			}
			contextLogger.Errorf("app: DeleteAllProductImages(ctx, productID=%q) failed: %v", productID, err)
			w.WriteHeader(http.StatusInternalServerError) // 500 Internal Server Error
			return
		}
		w.Header().Del("Content-Type")
		w.WriteHeader(http.StatusNoContent) // 204 No Content
	}
}
