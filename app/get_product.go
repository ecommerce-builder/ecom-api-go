package app

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"bitbucket.org/andyfusniakteam/ecom-api-go/model/postgres"
	"github.com/go-chi/chi"
)

// GetProductHandler returns a handler function that gets a product by SKU
// containing product and image data.
func (a *App) GetProductHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		sku := chi.URLParam(r, "sku")
		product, err := a.Service.GetProduct(r.Context(), sku)
		if err != nil {
			if err == postgres.ErrProductNotFound {
				w.WriteHeader(http.StatusNotFound)
				return
			}
			fmt.Fprintf(os.Stderr, "app: GetProduct(ctx, %q) error: %+v", sku, err)
			w.WriteHeader(http.StatusInternalServerError) // 500 Internal Server Error
			return
		}
		w.WriteHeader(http.StatusOK) // 200 OK
		json.NewEncoder(w).Encode(*product)
	}
}
