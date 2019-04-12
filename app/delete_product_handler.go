package app

import (
	"fmt"
	"net/http"
	"os"

	"github.com/go-chi/chi"
)

// DeleteProductHandler create a hanlder to delete a product resource.
func (a *App) DeleteProductHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		sku := chi.URLParam(r, "sku")

		err := a.Service.DeleteProduct(r.Context(), sku)
		if err != nil {
			fmt.Fprintf(os.Stderr, "delete product sku=%q failed: %v", sku, err)
			return
		}
		w.Header().Del("Content-Type")
		w.WriteHeader(http.StatusNoContent) // 204 No Content
	}
}
