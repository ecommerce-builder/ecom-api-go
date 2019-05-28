package app

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/go-chi/chi"
)

// ListProductImagesHandler creates a handler function that returns a
// slice of ProductImages for the product with the given SKU.
func (a *App) ListProductImagesHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		sku := chi.URLParam(r, "sku")
		products, err := a.Service.ListProductImages(r.Context(), sku)
		if err != nil {
			fmt.Fprintf(os.Stderr, "service ListProductImages(ctx, %s) error: %+v", sku, err)
			w.WriteHeader(http.StatusInternalServerError) // 500 Internal Server Error
			return
		}
		w.WriteHeader(http.StatusOK) // 200 OK
		json.NewEncoder(w).Encode(products)
	}
}
