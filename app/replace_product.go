package app

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	service "bitbucket.org/andyfusniakteam/ecom-api-go/service/firebase"
	"github.com/go-chi/chi"
)

func validateRequestBody(pc *service.ProductCreate) error {
	imagemap := make(map[string]bool)
	for _, img := range pc.Images {
		if _, found := imagemap[img.Path]; found {
			return fmt.Errorf("duplicate image path %s", img.Path)
		}
		imagemap[img.Path] = true
	}

	// TODO: make sure the new path is not already taken by another
	// product other than this one.
	return nil
}

// CreateReplaceProductHandler creates a new product if it does not exist, or
// updates and existing product.
//
// A separate call must be made to associate the product to the catalog
// hierarchy.
func (a *App) CreateReplaceProductHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		sku := chi.URLParam(r, "sku")
		pc := service.ProductCreate{}
		if err := json.NewDecoder(r.Body).Decode(&pc); err != nil {
			http.Error(w, err.Error(), 400)
			return
		}
		if err := validateRequestBody(&pc); err != nil {
			w.WriteHeader(http.StatusConflict) // 409 Conflict
			json.NewEncoder(w).Encode(struct {
				Status  int    `json:"status"`
				Code    string `json:"code"`
				Message string `json:"message"`
			}{
				http.StatusConflict,
				ErrCodeDuplicateImagePath,
				err.Error(),
			})
			return
		}
		product, err := a.Service.ReplaceProduct(r.Context(), sku, &pc)
		if err != nil {
			fmt.Fprintf(os.Stderr, "create product failed: %+v", err)
			w.WriteHeader(http.StatusInternalServerError) // 500 Internal Server Error
			return
		}
		w.WriteHeader(http.StatusCreated) // 201 Created
		json.NewEncoder(w).Encode(product)
	}
}
