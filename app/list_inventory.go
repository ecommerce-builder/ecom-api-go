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
			if err != nil {
				if err == postgres.ErrProductNotFound {
					clientError(w, http.StatusNotFound, ErrCodeProductNotFound, "product not found")
					return
				}

				contextLogger.Errorf("app: GetInventoryByProductID failed: %+v", err)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			w.WriteHeader(http.StatusOK) // 200 OK
			json.NewEncoder(w).Encode(&inventory)
			return
		}

		inventoryList, err := a.Service.GetAllInventory(ctx)
		if err != nil {
			if err == postgres.ErrInventoryNotFound {
				clientError(w, http.StatusNotFound, ErrCodeInventoryNotFound, "inventory not found")
				return
			}
			contextLogger.Errorf("app: a.Service.GetAllInventory(ctx) failed: %+v", err)
			w.WriteHeader(http.StatusInternalServerError)
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
