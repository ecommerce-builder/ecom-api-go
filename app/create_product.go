package app

import (
	"encoding/json"
	"net/http"

	service "bitbucket.org/andyfusniakteam/ecom-api-go/service/firebase"
	log "github.com/sirupsen/logrus"
)

// CreateProductHandler creates a new product
func (a *App) CreateProductHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		contextLogger := log.WithContext(ctx)
		contextLogger.Info("App: CreateProductHandler called")

		pc := service.ProductCreateUpdate{}
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
		defer r.Body.Close()

		product, err := a.Service.CreateUpdateProduct(ctx, nil, &pc)
		if err != nil {
			if err == service.ErrPricingTierNotFound {
				w.WriteHeader(http.StatusConflict) // 409 Conflict
				json.NewEncoder(w).Encode(struct {
					Status  int    `json:"status"`
					Code    string `json:"code"`
					Message string `json:"message"`
				}{
					http.StatusConflict,
					ErrCodePricingTierNotFound,
					"pricing tier could not be found",
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
