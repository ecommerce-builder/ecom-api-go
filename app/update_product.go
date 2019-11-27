package app

import (
	"encoding/json"
	"net/http"

	service "bitbucket.org/andyfusniakteam/ecom-api-go/service/firebase"
	"github.com/go-chi/chi"
	log "github.com/sirupsen/logrus"
)

func validateProductUpdateRequestBody(pc *service.ProductUpdateRequestBody) error {
	// imagemap := make(map[string]bool)
	// for _, img := range pc.Images {
	// 	if _, found := imagemap[img.Path]; found {
	// 		return fmt.Errorf("duplicate image path %s", img.Path)
	// 	}
	// 	imagemap[img.Path] = true
	// }

	// TODO: make sure the new path is not already taken by another
	// product other than this one.
	return nil
}

// UpdateProductHandler updates an existing product.
//
// A separate call must be made to associate the product to the catalog
// hierarchy.
func (a *App) UpdateProductHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		contextLogger := log.WithContext(ctx)
		contextLogger.Info("app: UpdateProductHandler started")

		productID := chi.URLParam(r, "id")
		pu := service.ProductUpdateRequestBody{}
		dec := json.NewDecoder(r.Body)
		dec.DisallowUnknownFields()
		if err := dec.Decode(&pu); err != nil {
			http.Error(w, err.Error(), 400)
			return
		}
		if err := validateProductUpdateRequestBody(&pu); err != nil {
			clientError(w, http.StatusBadRequest, ErrCodeBadRequest, err.Error())
			return
		}

		product, err := a.Service.UpdateProduct(ctx, productID, &pu)
		if err == service.ErrProductNotFound {
			clientError(w, http.StatusNotFound, ErrCodeProductNotFound,
				"product not found") // 404
			return
		}
		if err == service.ErrProductPathExists {
			clientError(w, http.StatusConflict, ErrCodeProductPathExists,
				"product path already exists") // 409
			return
		}
		if err == service.ErrProductSKUExists {
			clientError(w, http.StatusConflict, ErrCodeProductSKUExists,
				"product sku already exists") // 409
			return
		}
		if err != nil {
			contextLogger.Errorf("update product failed: %+v", err)
			w.WriteHeader(http.StatusInternalServerError) // 500
			return
		}
		w.WriteHeader(http.StatusOK) // 200
		json.NewEncoder(w).Encode(product)
	}
}
