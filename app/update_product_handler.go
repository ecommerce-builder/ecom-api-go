package app

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	service "bitbucket.org/andyfusniakteam/ecom-api-go/service/firebase"
	"github.com/go-chi/chi"
)

// UpdateProductHandler creates a handler to update a product by SKU.
func (a *App) UpdateProductHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		sku := chi.URLParam(r, "sku")
		pu := service.ProductUpdate{}
		err := json.NewDecoder(r.Body).Decode(&pu)
		if err != nil {
			http.Error(w, err.Error(), 400)
			return
		}
		product, err := a.Service.UpdateProduct(r.Context(), sku, &pu)
		if err != nil {
			fmt.Fprintf(os.Stderr, "update product failed: %+v", err)
			w.WriteHeader(http.StatusInternalServerError) // 500 Internal Server Error
			return
		}
		w.WriteHeader(http.StatusOK) // 200 OK
		json.NewEncoder(w).Encode(product)
	}
}
