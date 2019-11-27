package app

import (
	"encoding/json"
	"net/http"

	service "bitbucket.org/andyfusniakteam/ecom-api-go/service/firebase"
	"github.com/go-chi/chi"
	log "github.com/sirupsen/logrus"
)

type updateInventoryRequest struct {
	Onhand      *int  `json:"onhand"`
	Overselling *bool `json:"overselling"`
}

func validateUpdateInventoryRequest(request *updateInventoryRequest) (bool, string) {
	// onhand and overselling pair
	onhand := request.Onhand
	overselling := request.Overselling
	if onhand == nil && overselling == nil {
		return false, "you must set at least one attribute onhand and/or overselling"
	}

	if onhand != nil && *onhand < 0 {
		return false, "attribute onhand must be an positive integer or zero"
	}
	return true, ""
}

// UpdateInventoryHandler returns a http.HandlerFunc that returns the inventory
// by id.
func (a *App) UpdateInventoryHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		contextLogger := log.WithContext(ctx)
		contextLogger.Info("app: UpdateInventoryHandler called")

		// parse request body
		// example
		// { "onhand": 4, "overselling": true }
		var request updateInventoryRequest
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			clientError(w, http.StatusBadRequest, ErrCodeBadRequest, err.Error()) // 400
			return
		}

		ok, message := validateUpdateInventoryRequest(&request)
		if !ok {
			clientError(w, http.StatusBadRequest, ErrCodeBadRequest, message) // 400
			return
		}

		inventoryID := chi.URLParam(r, "id")
		if !IsValidUUID(inventoryID) {
			clientError(w, http.StatusBadRequest,
				ErrCodeBadRequest, "url parameter id must be a valid v4 UUID") // 400
			return
		}

		inventory, err := a.Service.UpdateInventory(ctx, inventoryID, request.Onhand, request.Overselling)
		if err == service.ErrInventoryNotFound {
			clientError(w, http.StatusNotFound, ErrCodeInventoryNotFound, "inventory not found") // 404
			return
		}
		if err != nil {
			contextLogger.Errorf("app: a.Service.UpdateInventory(ctx, inventoryID=%q, onhand=%v, overeslling=%v) failed: %+v", inventoryID, request.Onhand, request.Overselling, err)
			w.WriteHeader(http.StatusInternalServerError) // 500
			return
		}

		w.WriteHeader(http.StatusOK) // 200
		json.NewEncoder(w).Encode(&inventory)
	}
}
