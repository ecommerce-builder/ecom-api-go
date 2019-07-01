package app

import (
	"net/http"

	"github.com/go-chi/chi"
	log "github.com/sirupsen/logrus"
)

// DeleteAddressHandler deletes an address record
func (a *App) DeleteAddressHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		contextLogger := log.WithContext(ctx)
		contextLogger.Info("App: DeleteAddressHandler started")

		uuid := chi.URLParam(r, "uuid")
		err := a.Service.DeleteAddress(ctx, uuid)
		if err != nil {
			contextLogger.Errorf("service DeleteAddress(ctx, %s) failed with error: %v", uuid, err)
			return
		}
		w.Header().Del("Content-Type")
		w.WriteHeader(http.StatusNoContent) // 204 No Content
	}
}
