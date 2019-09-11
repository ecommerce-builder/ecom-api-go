package app

import (
	"encoding/json"
	"fmt"
	"net/http"

	service "bitbucket.org/andyfusniakteam/ecom-api-go/service/firebase"
	log "github.com/sirupsen/logrus"
)

type batchUpdateInventoryRequest struct {
	Object string                            `json:"object"`
	Data   []*service.InventoryUpdateRequest `json:"data"`
}

func validateBatchUpdateInventoryRequest(request *batchUpdateInventoryRequest) (bool, string) {
	if request.Object != "list" {
		return false, "object attribute must be set to list"
	}

	if request.Data == nil {
		return false, "data must be set to a list of inventory updates"
	}

	for _, update := range request.Data {
		if !IsValidUUID(update.ProductID) {
			return false, fmt.Sprintf("product %s not a valid uuid", update.ProductID)
		}
		if update.Onhand < 0 {
			return false, fmt.Sprintf("onhand must be a positive integer for product %s", update.ProductID)
		}
	}

	return true, ""
}

// BatchUpdateInventoryHandler returns a http.HandlerFunc that updates
// the inventory for a batch of products.
func (a *App) BatchUpdateInventoryHandler() http.HandlerFunc {
	type response struct {
		Object string               `json:"object"`
		Data   []*service.Inventory `json:"data"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		contextLogger := log.WithContext(ctx)
		contextLogger.Info("app: BatchUpdateInventoryHandler called")

		// parse the request body
		var request batchUpdateInventoryRequest
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			clientError(w, http.StatusBadRequest, ErrCodeBadRequest, err.Error())
			return
		}
		valid, message := validateBatchUpdateInventoryRequest(&request)
		if !valid {
			clientError(w, http.StatusBadRequest, ErrCodeBadRequest, message)
			return
		}

		// do the batch updates
		inventoryList, err := a.Service.BatchUpdateInventory(ctx, request.Data)
		if err != nil {
			if err == service.ErrProductNotFound {
				clientError(w, http.StatusNotFound, ErrCodeProductNotFound, "one or more products could not be found")
				return
			}

			contextLogger.Infof("app: a.Service.BatchUpdateInventory(ctx, request.Data) failed: %+v", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK) // 200 OK
		list := response{
			Object: "list",
			Data:   inventoryList,
		}
		json.NewEncoder(w).Encode(&list)
	}

}
