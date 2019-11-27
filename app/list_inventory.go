package app

import (
	"encoding/json"
	"net/http"

	"bitbucket.org/andyfusniakteam/ecom-api-go/model/postgres"
	service "bitbucket.org/andyfusniakteam/ecom-api-go/service/firebase"
	log "github.com/sirupsen/logrus"
)

// ListInventoryHandler returns a http.HandlerFunc that returns the inventory
// for a list of products.
func (a *App) ListInventoryHandler() http.HandlerFunc {
	type response struct {
		Object string               `json:"object"`
		Data   []*service.Inventory `json:"data"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		contextLogger := log.WithContext(ctx)
		contextLogger.Info("app: ListInventoryHandler called")

		productID := r.URL.Query().Get("product_id")
		if productID != "" {
			inventory, err := a.Service.GetInventoryByProductID(ctx, productID)
			if err == postgres.ErrProductNotFound {
				clientError(w, http.StatusNotFound, ErrCodeProductNotFound, "product not found") // 404
				return
			}
			if err != nil {
				contextLogger.Errorf("app: a.Service.GetInventoryByProductID(ctx, productID=%q) failed: %+v", productID, err)
				w.WriteHeader(http.StatusInternalServerError) // 500
				return
			}

			w.WriteHeader(http.StatusOK) // 200
			json.NewEncoder(w).Encode(&inventory)
			return
		}

		inventoryList, err := a.Service.GetAllInventory(ctx)
		if err == postgres.ErrInventoryNotFound {
			clientError(w, http.StatusNotFound, ErrCodeInventoryNotFound, "inventory not found") // 404
			return
		}
		if err != nil {
			contextLogger.Errorf("app: a.Service.GetAllInventory(ctx) failed: %+v", err)
			w.WriteHeader(http.StatusInternalServerError) // 500
			return
		}

		list := response{
			Object: "list",
			Data:   inventoryList,
		}
		w.WriteHeader(http.StatusOK) // 200 OK
		json.NewEncoder(w).Encode(&list)
	}
}
