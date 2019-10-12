package app

import (
	"encoding/json"
	"net/http"

	service "bitbucket.org/andyfusniakteam/ecom-api-go/service/firebase"
	"github.com/go-chi/chi"
	log "github.com/sirupsen/logrus"
)

// GetInventoryHandler returns a http.HandlerFunc that returns the inventory
// by id.
func (a *App) GetInventoryHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		contextLogger := log.WithContext(ctx)
		contextLogger.Info("app: GetInventoryHandler called")

		inventoryID := chi.URLParam(r, "id")
		if !IsValidUUID(inventoryID) {
			clientError(w, http.StatusBadRequest, ErrCodeBadRequest,
				"url parameter id must be a valid v4 UUID") // 400
			return
		}

		inventory, err := a.Service.GetInventory(ctx, inventoryID)
		if err == service.ErrInventoryNotFound {
			clientError(w, http.StatusNotFound, ErrCodeInventoryNotFound,
				"inventory not found") // 404
			return
		}
		if err != nil {
			contextLogger.Errorf("app: a.Service.GetInventory(ctx, inventoryID=%q) failed: %+v", inventoryID, err)
			w.WriteHeader(http.StatusInternalServerError) // 500
			return
		}

		w.WriteHeader(http.StatusOK) // 200
		json.NewEncoder(w).Encode(&inventory)
	}
}
