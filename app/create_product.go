package app

import (
	"encoding/json"
	"net/http"

	service "bitbucket.org/andyfusniakteam/ecom-api-go/service/firebase"
	log "github.com/sirupsen/logrus"
)

func validateProductCreateRequestBody(pc *service.ProductCreateRequestBody) (bool, string) {
	return true, ""
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
		defer r.Body.Close()

		valid, message := validateProductCreateRequestBody(&pc)
		if !valid {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(struct {
				Status  int    `json:"status"`
				Code    string `json:"code"`
				Message string `json:"message"`
			}{
				http.StatusBadRequest,
				ErrCodeBadRequest,
				message,
			})
			return
		}

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
					"product SKU already exists",
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
