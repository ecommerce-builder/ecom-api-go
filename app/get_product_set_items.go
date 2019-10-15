package app

import (
	"encoding/json"
	"net/http"

	service "bitbucket.org/andyfusniakteam/ecom-api-go/service/firebase"
	log "github.com/sirupsen/logrus"
)

// GetProductSetItemsHandler returns a list of product set items.
func (a *App) GetProductSetItemsHandler() http.HandlerFunc {
	type responseBody struct {
		Object string                    `json:"object"`
		Data   []*service.ProductSetItem `json:"data"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		contextLogger := log.WithContext(ctx)
		contextLogger.Info("app: GetProductSetItemsHandler started")

		productSetID := r.URL.Query().Get("product_set_id")
		if !IsValidUUID(productSetID) {
			clientError(w, http.StatusBadRequest, ErrCodeBadRequest, "product_set_id is not a valid UUID v4")
			return
		}

		productSetItems, err := a.Service.GetProductSetItems(ctx, productSetID)
		if err == service.ErrProductSetNotFound {
			w.WriteHeader(http.StatusNotFound)
			clientError(w, http.StatusNotFound, ErrCodeProductSetNotFound,
				"product set not found") // 404
			return
		}
		if err != nil {
			contextLogger.Errorf("app: a.Service.GetProductSetItems(ctx, productSetUUID=%q) failed: %+v", productSetID, err)
			w.WriteHeader(http.StatusInternalServerError) // 500
			return
		}
		response := responseBody{
			Object: "list",
			Data:   productSetItems,
		}
		w.WriteHeader(http.StatusOK) // 200 OK
		json.NewEncoder(w).Encode(&response)
	}
}
