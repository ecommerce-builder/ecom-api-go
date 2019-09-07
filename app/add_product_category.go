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
		contextLogger.Info("app: AddProductCategoryHandler called")

		request := service.ProductCategoryRequestBody{}
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			clientError(w, http.StatusBadRequest, ErrCodeBadRequest, err.Error())
			return
		}

		productCategory, err := a.Service.AddProductCategory(ctx, request.ProductID, request.CategoryID)
		if err != nil {
			if err == service.ErrCategoryNotFound {
				clientError(w, http.StatusNotFound, ErrCodeCategoryNotFound, "category not found")
				return
			} else if err == service.ErrCategoryNotLeaf {
				clientError(w, http.StatusConflict, ErrCodeCategoryNotLeaf, "category is not a leaf. products can only be associated to leaf categories")
				return
			} else if err == service.ErrProductNotFound {
				clientError(w, http.StatusNotFound, ErrCodeProductNotFound, "product not found")
				return
			} else if err == service.ErrProductCategoryExists {
				clientError(w, http.StatusConflict, ErrCodeProductCategoryExists, "product to category association already exists")
				return
			}

			contextLogger.Errorf("a.Service.AddProductCategory(ctx, productID=%q, categoryID=%q) failed with error: %+v", request.ProductID, request.CategoryID, err)
			w.WriteHeader(http.StatusInternalServerError) // 500 Internal Server Error
			return
		}
		w.WriteHeader(http.StatusCreated) // 201 Created
		json.NewEncoder(w).Encode(&productCategory)
	}
}
