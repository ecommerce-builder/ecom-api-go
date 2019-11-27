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
		// product_id attribute
		productID := update.ProductID
		if productID == nil {
			return false, "attribute product_id must be set for all items"
		}
		if !IsValidUUID(*productID) {
			return false, fmt.Sprintf("product %s not a valid uuid", *update.ProductID)
		}

		// onhand attribute
		onhand := update.Onhand
		if onhand == nil {
			return false, "attribute onhand must be set for all items"
		}
		if *onhand < 0 {
			return false, fmt.Sprintf("onhand must be a positive integer for product=%q", *update.ProductID)
		}

		// overselling attribute
		overselling := update.Overselling
		if overselling == nil {
			return false, "attribute overselling must be set for all items"
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
		dec := json.NewDecoder(r.Body)
		dec.DisallowUnknownFields()
		if err := dec.Decode(&request); err != nil {
			clientError(w, http.StatusBadRequest, ErrCodeBadRequest, err.Error()) // 400
			return
		}
		valid, message := validateBatchUpdateInventoryRequest(&request)
		if !valid {
			clientError(w, http.StatusBadRequest, ErrCodeBadRequest, message) // 400
			return
		}

		// do the batch updates
		inventoryList, err := a.Service.BatchUpdateInventory(ctx, request.Data)
		if err == service.ErrProductNotFound {
			clientError(w, http.StatusNotFound,
				ErrCodeProductNotFound, "one or more products could not be found") // 404
			return
		}
		if err != nil {
			contextLogger.Infof("app: a.Service.BatchUpdateInventory(ctx, request.Data) failed: %+v", err)
			w.WriteHeader(http.StatusInternalServerError) // 500
			return
		}

		w.WriteHeader(http.StatusOK) // 200
		list := response{
			Object: "list",
			Data:   inventoryList,
		}
		json.NewEncoder(w).Encode(&list)
	}

}
