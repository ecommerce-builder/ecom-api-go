package app

import (
	"fmt"
	"net/http"
	"os"

	"github.com/go-chi/chi"
)

// ProductExistsHandler returns a HandlerFunc that checks to see if a
// product resource exists.
func (app *App) ProductExistsHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		sku := chi.URLParam(r, "sku")
		exists, err := app.Service.ProductExists(r.Context(), sku)
		if err != nil {
			fmt.Fprintf(os.Stderr, "product exists failed for sku=%q: %v", sku, err)
			return
		}
		if !exists {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.WriteHeader(http.StatusNoContent) // 204 No Content
	}
}
