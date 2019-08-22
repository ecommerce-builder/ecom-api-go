package app

import (
	"encoding/json"
	"net/http"

	service "bitbucket.org/andyfusniakteam/ecom-api-go/service/firebase"
	log "github.com/sirupsen/logrus"
)

// ListProductsHandler creates a handler that returns a list of products.
func (a *App) ListProductsHandler() http.HandlerFunc {
	type itemsListResponseBody struct {
		Object string             `json:"object"`
		Data   []*service.Product `json:"data"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		contextLogger := log.WithContext(ctx)
		contextLogger.Info("App: ListProductsHandler started")

		products, err := a.Service.ListProducts(ctx)
		if err != nil {
			contextLogger.Errorf("service: ListProducts(ctx) error: %v", err)
			w.WriteHeader(http.StatusInternalServerError) // 500 Internal Server Error
			return
		}
		res := itemsListResponseBody{
			Object: "list",
			Data:   products,
		}
		w.WriteHeader(http.StatusOK) // 200 OK
		json.NewEncoder(w).Encode(&res)
	}
}
