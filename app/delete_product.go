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

		productID := chi.URLParam(r, "id")
		if err := a.Service.DeleteProduct(ctx, productID); err != nil {
			contextLogger.Errorf("a.Service.DeleteProduct(ctx, productID=%q) failed: %v", productID, err)
			w.WriteHeader(http.StatusInternalServerError) // 500 Internal Server Error
			return
		}
		w.Header().Del("Content-Type")
		w.Header().Set("Content-Length", "0")
		w.WriteHeader(http.StatusNoContent) // 204 No Content
	}
}
