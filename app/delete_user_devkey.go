package app

import (
	"net/http"

	service "bitbucket.org/andyfusniakteam/ecom-api-go/service/firebase"
	"github.com/go-chi/chi"
	log "github.com/sirupsen/logrus"
)

// DeleteUserDevKeyHandler deletes a customer Developer Key.
func (a *App) DeleteUserDevKeyHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		contextLogger := log.WithContext(ctx)
		contextLogger.Info("app: DeleteUserDevKeyHandler started")

		developerKeyID := chi.URLParam(r, "id")
		if err := a.Service.DeleteDeveloperKey(ctx, developerKeyID); err != nil {
			if err == service.ErrDeveloperKeyNotFound {
				clientError(w, http.StatusNotFound, ErrCodeDeveloperKeyNotFound, "developer key not found")
				return
			}
		}
		contextLogger.Infof("app: Developer Key %s deleted", developerKeyID)
		w.WriteHeader(http.StatusNoContent) // 204 No Content
		return
	}
}
