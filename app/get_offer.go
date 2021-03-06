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

		offerID := chi.URLParam(r, "id")
		if !IsValidUUID(offerID) {
			contextLogger.Warn("400 Bad Request - path parameter id must be a valid v4 uuid")
			clientError(w, http.StatusBadRequest, ErrCodeBadRequest,
				"path parameter id must be a valid v4 uuid") // 400
			return
		}
		offer, err := a.Service.GetOffer(ctx, offerID)
		if err == service.ErrOfferNotFound {
			contextLogger.Warnf("404 Not Found - offer %s not found", offerID)
			clientError(w, http.StatusNotFound, ErrCodeOfferNotFound,
				"offer not found") // 404
			return
		}
		if err != nil {
			contextLogger.Errorf("app: a.Service.GetInventory(ctx, inventoryUUID=%q) failed: %+v",
				offerID, err)
			w.WriteHeader(http.StatusInternalServerError) // 500
			return
		}
		w.WriteHeader(http.StatusOK) // 200
		json.NewEncoder(w).Encode(&offer)
	}
}
