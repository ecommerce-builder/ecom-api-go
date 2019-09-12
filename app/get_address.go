package app

import (
	"encoding/json"
	"net/http"

	service "bitbucket.org/andyfusniakteam/ecom-api-go/service/firebase"
	"github.com/go-chi/chi"
	log "github.com/sirupsen/logrus"
)

// GetAddressHandler gets an address by its address UUID
func (app *App) GetAddressHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		contextLogger := log.WithContext(ctx)
		contextLogger.Info("app: GetAddressHandler started")

		addressID := chi.URLParam(r, "id")
		if addressID == "" {
			contextLogger.Warn("app: URL param id not set")
			clientError(w, http.StatusBadRequest, ErrCodeBadRequest, "URL parameter id must be set")
			return
		} else if !IsValidUUID(addressID) {
			contextLogger.Warn("app: URL param is not a valid v4 UUID")
			clientError(w, http.StatusBadRequest, ErrCodeBadRequest, "URL parameter id must be set to a valid v4 UUID")
			return
		}

		addr, err := app.Service.GetAddress(ctx, addressID)
		if err != nil {
			if err == service.ErrAddressNotFound {
				contextLogger.Warnf("app: address %q not found: %v", addressID, err)
				clientError(w, http.StatusNotFound, ErrCodeAddressNotFound, "address not found")
				return
			}
			contextLogger.Errorf("app: failed to get address: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK) // 200 OK
		json.NewEncoder(w).Encode(&addr)
	}
}
