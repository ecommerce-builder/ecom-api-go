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
		contextLogger.Info("app: DeleteAddressHandler started")

		id := chi.URLParam(r, "id")
		err := a.Service.DeleteAddress(ctx, id)
		if err != nil {
			contextLogger.Errorf("app: DeleteAddress(ctx, %s) failed with error: %v", id, err)
			return
		}
		w.Header().Del("Content-Type")
		w.WriteHeader(http.StatusNoContent) // 204 No Content
	}
}
