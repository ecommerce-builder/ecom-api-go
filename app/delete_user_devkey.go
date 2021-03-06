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
		if !IsValidUUID(developerKeyID) {
			contextLogger.Warnf("app: 400 Bad Request - path parameter must be a valid v4 uuid (got %q)",
				developerKeyID)
			clientError(w, http.StatusBadRequest, ErrCodeBadRequest,
				"path parameters id must be a valid v4 uuid") // 400
			return
		}

		err := a.Service.DeleteDeveloperKey(ctx, developerKeyID)
		if err == service.ErrDeveloperKeyNotFound {
			contextLogger.Warnf("app: 404 Not Found - developer key id %q not found",
				developerKeyID)
			clientError(w, http.StatusNotFound, ErrCodeDeveloperKeyNotFound,
				"developer key not found") // 404
			return
		}
		if err != nil {
			contextLogger.Errorf("app: a.Service.DeleteDeveloperKey(ctx, developerKeyID=%q) failed: %+v", developerKeyID, err)
			w.WriteHeader(http.StatusInternalServerError) // 500
			return
		}
		contextLogger.Infof("app: Developer Key %s deleted", developerKeyID)
		w.WriteHeader(http.StatusNoContent) // 204
		return
	}
}
