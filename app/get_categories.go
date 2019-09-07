package app

import (
	"encoding/json"
	"net/http"

	service "bitbucket.org/andyfusniakteam/ecom-api-go/service/firebase"
	log "github.com/sirupsen/logrus"
)

// GetCategoriesHandler returns a handler function that returns categories
// in nested set form.
func (a *App) GetCategoriesHandler() http.HandlerFunc {
	type listResponse struct {
		Object string              `json:"object"`
		Data   []*service.Category `json:"data"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		contextLogger := log.WithContext(ctx)
		contextLogger.Info("app: GetCategoriesHandler called")

		categories, err := a.Service.GetCategories(ctx)
		if err != nil {

			contextLogger.Errorf("app: a.Service.GetCategories(ctx) error: %+v", err)
			w.WriteHeader(http.StatusInternalServerError) // 500 Internal Server Error
			return
		}

		list := listResponse{
			Object: "list",
			Data:   categories,
		}
		w.WriteHeader(http.StatusOK) // 200 OK
		json.NewEncoder(w).Encode(&list)
	}
}
