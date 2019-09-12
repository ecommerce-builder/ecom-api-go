package app

import (
	"net/http"

	service "bitbucket.org/andyfusniakteam/ecom-api-go/service/firebase"
	"github.com/go-chi/chi"
	log "github.com/sirupsen/logrus"
)

// DeleteAddressHandler returns a http.HandlerFunc that attempts to
// delete an address.
func (a *App) DeleteAddressHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		contextLogger := log.WithContext(ctx)
		contextLogger.Info("app: DeleteAddressHandler started")

		addressID := chi.URLParam(r, "id")
		if !IsValidUUID(addressID) {
			clientError(w, http.StatusBadRequest, ErrCodeBadRequest, "URL parameter id must be a valid v4 UUID")
			return
		}

		err := a.Service.DeleteAddress(ctx, addressID)
		if err != nil {
			if err == service.ErrAddressNotFound {
				clientError(w, http.StatusNotFound, ErrCodeAddressNotFound, "address not found")
				return
			}
			contextLogger.Errorf("app: DeleteAddress(ctx, addressID=%q) failed with error: %v", addressID, err)
			return
		}

		contextLogger.Infof("app: address %q delete", addressID)
		w.Header().Del("Content-Type")
		w.WriteHeader(http.StatusNoContent) // 204 No Content
	}
}
