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
		inventory, err := a.Service.GetInventory(ctx, inventoryID)
		if err != nil {
			if err == service.ErrInventoryNotFound {
				clientError(w, http.StatusNotFound, ErrCodeInventoryNotFound, "inventory not found")
				return
			}
			contextLogger.Errorf("app: a.Service.GetInventory(ctx, inventoryUUID=%q) failed: %+v", inventoryID, err)
			return
		}

		w.WriteHeader(http.StatusOK) // 200 OK
		json.NewEncoder(w).Encode(&inventory)
	}
}
