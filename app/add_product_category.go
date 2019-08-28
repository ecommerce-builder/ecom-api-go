package app

import (
	"encoding/json"
	"net/http"

	service "bitbucket.org/andyfusniakteam/ecom-api-go/service/firebase"
	log "github.com/sirupsen/logrus"
)

// AddProductCategoryHandler links a product to a leaf node category
func (a *App) AddProductCategoryHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		contextLogger := log.WithContext(ctx)
		contextLogger.Info("App: AddProductCategoryHandler called")

		requestBody := service.ProductCategoryRequestBody{}
		if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(struct {
				Status  int    `json:"status"`
				Code    string `json:"code"`
				Message string `json:"message"`
			}{
				http.StatusBadRequest,
				ErrCodeBadRequest,
				"bad request",
			})
			return
		}

		productCategory, err := a.Service.AddProductCategory(ctx, &requestBody)
		if err != nil {
			if err == service.ErrCategoryNotFound {
				w.WriteHeader(http.StatusConflict)
				json.NewEncoder(w).Encode(struct {
					Status  int    `json:"status"`
					Code    string `json:"code"`
					Message string `json:"message"`
				}{
					http.StatusConflict,
					ErrCodeCategoryNotFound,
					"category not found",
				})
				return
			} else if err == service.ErrCategoryNotLeaf {
				w.WriteHeader(http.StatusConflict)
				json.NewEncoder(w).Encode(struct {
					Status  int    `json:"status"`
					Code    string `json:"code"`
					Message string `json:"message"`
				}{
					http.StatusConflict,
					ErrCodeCategoryNotLeaf,
					"category is not a leaf. products can only be associated to leaf categories",
				})
				return
			} else if err == service.ErrProductNotFound {
				w.WriteHeader(http.StatusConflict)
				json.NewEncoder(w).Encode(struct {
					Status  int    `json:"status"`
					Code    string `json:"code"`
					Message string `json:"message"`
				}{
					http.StatusConflict,
					ErrCodeProductNotFound,
					"product not found",
				})
				return
			}

			contextLogger.Errorf("a.Service.AddProductCategory(ctx, request=%v) failed with error: %+v", requestBody, err)
			w.WriteHeader(http.StatusInternalServerError) // 500 Internal Server Error
			return
			// 	if err == service.ErrPriceListCodeExists {
			// 		w.WriteHeader(http.StatusConflict) // 409 Conflict
			// 		json.NewEncoder(w).Encode(struct {
			// 			Status  int    `json:"status"`
			// 			Code    string `json:"code"`
			// 			Message string `json:"message"`
			// 		}{
			// 			http.StatusConflict,
			// 			ErrCodePriceListCodeExists,
			// 			"price list is already in use",
			// 		})
			// 		return
			// 	}
		}
		w.WriteHeader(http.StatusCreated) // 201 Created
		json.NewEncoder(w).Encode(&productCategory)
	}
}
