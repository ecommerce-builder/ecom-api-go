package app

import (
	"net/http"

	"github.com/go-chi/chi"
	log "github.com/sirupsen/logrus"
)

// ProductExistsHandler returns a HandlerFunc that checks to see if a
// product resource exists.
func (app *App) ProductExistsHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		contextLogger := log.WithContext(ctx)
		contextLogger.Info("App: ProductExistsHandler started")

		sku := chi.URLParam(r, "sku")
		exists, err := app.Service.ProductExists(ctx, sku)
		if err != nil {
			contextLogger.Errorf("product exists failed for sku=%q: %v", sku, err)
			w.WriteHeader(http.StatusInternalServerError) // 500 Internal Server Error
			return
		}
		if !exists {
			w.WriteHeader(http.StatusNotFound) // 404 Not Found
			return
		}
		w.WriteHeader(http.StatusNoContent) // 204 No Content
	}
}
