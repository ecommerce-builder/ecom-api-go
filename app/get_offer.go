package app

import (
	"encoding/json"
	"net/http"

	service "bitbucket.org/andyfusniakteam/ecom-api-go/service/firebase"
	"github.com/go-chi/chi"
	log "github.com/sirupsen/logrus"
)

// GetOfferHandler returns a http.HandlerFunc that returns an offer
// by id.
func (a *App) GetOfferHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		contextLogger := log.WithContext(ctx)
		contextLogger.Info("app: GetOfferHandler called")

		inventoryID := chi.URLParam(r, "id")
		inventory, err := a.Service.GetOffer(ctx, inventoryID)
		if err != nil {
			if err == service.ErrInventoryNotFound {
				clientError(w, http.StatusNotFound, ErrCodeOfferNotFound, "offer not found")
				return
			}
			contextLogger.Errorf("app: a.Service.GetInventory(ctx, inventoryUUID=%q) failed: %+v", inventoryID, err)
			return
		}

		w.WriteHeader(http.StatusOK) // 200 OK
		json.NewEncoder(w).Encode(&inventory)
	}
}
