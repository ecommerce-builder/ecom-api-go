package app

import (
	"encoding/json"
	"net/http"

	service "bitbucket.org/andyfusniakteam/ecom-api-go/service/firebase"
	"github.com/go-chi/chi"
	log "github.com/sirupsen/logrus"
)

type updateInventoryRequest struct {
	Onhold int `json:"onhold"`
}

func validateUpdateInventoryRequest(request *updateInventoryRequest) (bool, string) {
	if request.Onhold < 0 {
		return false, ""
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
		// { "onhold": 4 }
		var request updateInventoryRequest
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			clientError(w, http.StatusBadRequest, ErrCodeBadRequest, err.Error())
			return
		}

		ok, message := validateUpdateInventoryRequest(&request)
		if !ok {
			clientError(w, http.StatusBadRequest, ErrCodeBadRequest, message)
			return
		}

		inventoryID := chi.URLParam(r, "id")
		inventory, err := a.Service.UpdateInventory(ctx, inventoryID, request.Onhold)
		if err != nil {
			if err == service.ErrInventoryNotFound {
				clientError(w, http.StatusNotFound, ErrCodeInventoryNotFound, "inventory not found")
				return
			}
			contextLogger.Errorf("app: a.Service.GetInventory(ctx, inventoryUUID=%q) failed: %+v", inventoryID, err)
			w.WriteHeader(http.StatusInternalServerError) // 500 Internal Server Error
			return
		}

		w.WriteHeader(http.StatusOK) // 200 OK
		json.NewEncoder(w).Encode(&inventory)
	}
}
