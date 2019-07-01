package app

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi"
	log "github.com/sirupsen/logrus"
)

// ListProductImagesHandler creates a handler function that returns a
// slice of ProductImages for the product with the given SKU.
func (a *App) ListProductImagesHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		contextLogger := log.WithContext(ctx)
		contextLogger.Info("App: ListProductImagesHandler started")

		sku := chi.URLParam(r, "sku")
		products, err := a.Service.ListProductImages(ctx, sku)
		if err != nil {
			contextLogger.Errorf("service ListProductImages(ctx, %s) error: %+v", sku, err)
			w.WriteHeader(http.StatusInternalServerError) // 500 Internal Server Error
			return
		}
		w.WriteHeader(http.StatusOK) // 200 OK
		json.NewEncoder(w).Encode(products)
	}
}
