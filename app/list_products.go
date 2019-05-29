package app

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
)

// ListProductsHandler creates a handler that returns a list of products.
func (a *App) ListProductsHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		products, err := a.Service.ListProducts(r.Context())
		if err != nil {
			fmt.Fprintf(os.Stderr, "service: ListProducts(ctx) error: %v", err)
			w.WriteHeader(http.StatusInternalServerError) // 500 Internal Server Error
			return
		}
		w.WriteHeader(http.StatusOK) // 200 OK
		json.NewEncoder(w).Encode(products)
	}
}
