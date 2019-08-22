package app

import (
	"encoding/json"
	"net/http"

	service "bitbucket.org/andyfusniakteam/ecom-api-go/service/firebase"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

// UpdateProductsCategoriesHandler returns a handler to batch update product to categories.
func (app *App) UpdateProductsCategoriesHandler() http.HandlerFunc {
	type request struct {
		Object string                              `json:"object"`
		Data   []*service.CreateProductsCategories `json:"data"`
	}

	type response struct {
		Object string                        `json:"object"`
		Data   []*service.ProductsCategories `json:"data"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		contextLogger := log.WithContext(ctx)
		contextLogger.Info("App: UpdateProductsCategoriesHandler called")

		var request request
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			json.NewEncoder(w).Encode(struct {
				Status  int    `json:"status"`
				Code    string `json:"code"`
				Message string `json:"message"`
			}{
				http.StatusBadRequest,
				ErrCodeBadRequest,
				err.Error(),
			})
			return
		}

		productsCategories, err := app.Service.UpdateProductsCategories(ctx, request.Data)
		if err != nil {
			if err == service.ErrProductNotFound {
				w.WriteHeader(http.StatusConflict) // 409 Conflict
				json.NewEncoder(w).Encode(struct {
					Status  int    `json:"status"`
					Code    string `json:"code"`
					Message string `json:"message"`
				}{
					http.StatusConflict,
					ErrCodeProductNotFound,
					"one or more product ids cannot be found",
				})
				return
			} else if err == service.ErrLeafCategoryNotFound {
				w.WriteHeader(http.StatusConflict) // 409 Conflict
				json.NewEncoder(w).Encode(struct {
					Status  int    `json:"status"`
					Code    string `json:"code"`
					Message string `json:"message"`
				}{
					http.StatusConflict,
					ErrCodeLeafCategoryNotFound,
					"one or more leaf category ids cannot be found",
				})
				return
			}
			contextLogger.Errorf("service UpdateProductsCategories(ctx, %v) error: %+v", request, errors.WithStack(err))
			w.WriteHeader(http.StatusInternalServerError) // 500
			return
		}

		list := response{
			Object: "list",
			Data:   productsCategories,
		}
		w.WriteHeader(http.StatusOK) // 200 OK
		if err = json.NewEncoder(w).Encode(&list); err != nil {
			contextLogger.Errorf("json encode failed with error: %+v", err)
			w.WriteHeader(http.StatusInternalServerError) // 500 Internal Server Error
			return
		}
	}
}
