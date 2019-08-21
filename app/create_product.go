package app

import (
	"encoding/json"
	"fmt"
	"net/http"

	service "bitbucket.org/andyfusniakteam/ecom-api-go/service/firebase"
	log "github.com/sirupsen/logrus"
)

func validateProductCreateRequestBody(pc *service.ProductCreateRequestBody) error {
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

// CreateProductHandler creates a new product
func (a *App) CreateProductHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		contextLogger := log.WithContext(ctx)
		contextLogger.Info("app: CreateProductHandler called")

		pc := service.ProductCreateRequestBody{}
		if err := json.NewDecoder(r.Body).Decode(&pc); err != nil {
			http.Error(w, err.Error(), 400)
			return
		}

		if err := validateProductCreateRequestBody(&pc); err != nil {
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
		defer r.Body.Close()

		product, err := a.Service.CreateProduct(ctx, &pc)
		if err != nil {
			if err == service.ErrPriceListNotFound {
				w.WriteHeader(http.StatusConflict) // 409 Conflict
				json.NewEncoder(w).Encode(struct {
					Status  int    `json:"status"`
					Code    string `json:"code"`
					Message string `json:"message"`
				}{
					http.StatusConflict,
					ErrCodePriceListNotFound,
					"price list could not be found",
				})
				return
			}
			contextLogger.Errorf("create product failed: %+v", err)
			w.WriteHeader(http.StatusInternalServerError) // 500 Internal Server Error
			return
		}
		w.WriteHeader(http.StatusCreated) // 201 Created
		json.NewEncoder(w).Encode(product)
	}
}
