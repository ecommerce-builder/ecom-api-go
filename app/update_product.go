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
		if err := json.NewDecoder(r.Body).Decode(&pu); err != nil {
			http.Error(w, err.Error(), 400)
			return
		}
		if err := validateProductUpdateRequestBody(&pu); err != nil {
			w.WriteHeader(http.StatusBadRequest) // 400 Bad Requst
			json.NewEncoder(w).Encode(struct {
				Status  int    `json:"status"`
				Code    string `json:"code"`
				Message string `json:"message"`
			}{
				http.StatusConflict,
				ErrCodeBadRequest,
				err.Error(),
			})
			return
		}

		product, err := a.Service.UpdateProduct(ctx, productID, &pu)
		if err != nil {
			if err == service.ErrProductNotFound {
				w.WriteHeader(http.StatusNotFound) // 404 Not Found
				json.NewEncoder(w).Encode(struct {
					Status  int    `json:"status"`
					Code    string `json:"code"`
					Message string `json:"message"`
				}{
					http.StatusNotFound,
					ErrCodeProductNotFound,
					"product not found",
				})
				return
			} else if err == service.ErrProductPathExists {
				w.WriteHeader(http.StatusConflict) // 409 Conflict
				json.NewEncoder(w).Encode(struct {
					Status  int    `json:"status"`
					Code    string `json:"code"`
					Message string `json:"message"`
				}{
					http.StatusConflict,
					ErrCodeProductPathExists,
					"product path already exists",
				})
				return
			} else if err == service.ErrProductSKUExists {
				w.WriteHeader(http.StatusConflict) // 409 Conflict
				json.NewEncoder(w).Encode(struct {
					Status  int    `json:"status"`
					Code    string `json:"code"`
					Message string `json:"message"`
				}{
					http.StatusConflict,
					ErrCodeProductSKUExists,
					"product sku already exists",
				})
				return
			}
			contextLogger.Errorf("update product failed: %+v", err)
			w.WriteHeader(http.StatusInternalServerError) // 500 Internal Server Error
			return
		}
		w.WriteHeader(http.StatusOK) // 200 OK
		json.NewEncoder(w).Encode(product)
	}
}
