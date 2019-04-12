package app

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
)

// CreateProductHandler create a new product. A separate call must be made
// to associate the product to the catalog hierarchy.
func (a *App) CreateProductHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		pc := ProductCreate{}
		err := json.NewDecoder(r.Body).Decode(&pc)
		if err != nil {
			http.Error(w, err.Error(), 400)
			return
		}

		product, err := a.Service.CreateProduct(r.Context(), &pc)
		if err != nil {
			fmt.Fprintf(os.Stderr, "create product failed: %v", err)
			return
		}
		w.WriteHeader(http.StatusCreated) // 201 Created
		json.NewEncoder(w).Encode(product)
	}
}
