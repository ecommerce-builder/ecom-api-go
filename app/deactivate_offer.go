package app

import (
	"net/http"

	service "bitbucket.org/andyfusniakteam/ecom-api-go/service/firebase"
	"github.com/go-chi/chi"
	log "github.com/sirupsen/logrus"
)

// DeactivateOfferHandler creates a handler function that deletes
// a price list by id.
func (a *App) DeactivateOfferHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		contextLogger := log.WithContext(ctx)
		contextLogger.Info("app: DeactivateOfferHandler started")

		offerID := chi.URLParam(r, "id")
		err := a.Service.DeactivateOffer(ctx, offerID)
		if err == service.ErrOfferNotFound {
			clientError(w, http.StatusNotFound, ErrCodeOfferNotFound, "offer not found")
			return
		}
		if err != nil {
			contextLogger.Errorf("app: a.Service.DeactivateOffer(ctx, offerID=%q) failed: %+v", offerID, err)
			w.WriteHeader(http.StatusInternalServerError) // 500 Internal Server Error
			return
		}
		w.Header().Del("Content-Type")
		w.WriteHeader(http.StatusNoContent) // 204 No Content
	}
}
