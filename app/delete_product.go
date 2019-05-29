package app

import (
	"fmt"
	"net/http"
	"os"

	"github.com/go-chi/chi"
)

// DeleteProductHandler create a handler to delete a product resource.
func (a *App) DeleteProductHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		sku := chi.URLParam(r, "sku")
		if err := a.Service.DeleteProduct(r.Context(), sku); err != nil {
			fmt.Fprintf(os.Stderr, "delete product sku=%q failed: %v", sku, err)
			w.WriteHeader(http.StatusInternalServerError) // 500 Internal Server Error
			return
		}
		w.Header().Del("Content-Type")
		w.WriteHeader(http.StatusNoContent) // 204 No Content
	}
}
