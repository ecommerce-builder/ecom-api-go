package app

import (
	"encoding/json"
	"net/http"

	service "bitbucket.org/andyfusniakteam/ecom-api-go/service/firebase"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

// GetProductsCategoriesHandler creates a handler to return product to category associations.
func (app *App) GetProductsCategoriesHandler() http.HandlerFunc {
	type productsCategoriesListResponse struct {
		Object string                        `json:"object"`
		Data   []*service.ProductsCategories `json:"data"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		contextLogger := log.WithContext(ctx)
		contextLogger.Info("App: GetProductsCategoriesHandler called")

		assocs, err := app.Service.GetProductsCategoriesList(ctx)
		if err != nil {
			contextLogger.Errorf("service GetProductsCategoriesList(ctx) error: %+v", errors.WithStack(err))
			w.WriteHeader(http.StatusInternalServerError) // 500
			return
		}

		list := productsCategoriesListResponse{
			Object: "list",
			Data:   assocs,
		}

		w.WriteHeader(http.StatusOK) // 200 OK
		if err = json.NewEncoder(w).Encode(&list); err != nil {
			contextLogger.Errorf("app: json encode failed with error: %+v", err)
			w.WriteHeader(http.StatusInternalServerError) // 500 Internal Server Error
			return
		}
	}
}
