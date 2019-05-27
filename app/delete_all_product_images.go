package app

import (
	"fmt"
	"net/http"
	"os"

	"github.com/go-chi/chi"
)

// DeleteAllProductImagesHandler create a handler that deletes all images for the
// product with the given SKU.
func (a *App) DeleteAllProductImagesHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		sku := chi.URLParam(r, "sku")
		exists, err := a.Service.ProductExists(r.Context(), sku)
		if err != nil {
			fmt.Fprintf(os.Stderr, "ProductExists(ctx, %q) failed: %v", sku, err)
			w.WriteHeader(http.StatusInternalServerError) // 500 Internal Server Error
			return
		}
		if !exists {
			w.WriteHeader(http.StatusNotFound) // 404 Not Found
			return
		}
		if err := a.Service.DeleteAllProductImages(r.Context(), sku); err != nil {
			fmt.Fprintf(os.Stderr, "DeleteAllProductImages(ctx, %q) failed: %v", sku, err)
			w.WriteHeader(http.StatusInternalServerError) // 500 Internal Server Error
			return
		}
		w.Header().Del("Content-Type")
		w.WriteHeader(http.StatusNoContent) // 204 No Content
	}
}
