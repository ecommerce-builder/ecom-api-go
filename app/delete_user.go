package app

import (
	"net/http"

	service "bitbucket.org/andyfusniakteam/ecom-api-go/service/firebase"
	"github.com/go-chi/chi"
	log "github.com/sirupsen/logrus"
)

// DeleteUserHandler creates an HTTP handler that deletes a user.
func (a *App) DeleteUserHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		contextLogger := log.WithContext(ctx)
		contextLogger.Info("app: DeleteUserHandler started")

		userID := chi.URLParam(r, "id")
		if !IsValidUUID(userID) {
			clientError(w, http.StatusBadRequest, ErrCodeBadRequest,
				"URL parameter id must be a valid v4 UUID") // 400
			return
		}

		// Attempt to delete this user.
		err := a.Service.DeleteUser(ctx, userID)
		if err == service.ErrUserNotFound {
			clientError(w, http.StatusNotFound, ErrCodeUserNotFound, "user not found") // 404
			return
		}
		if err == service.ErrUserInUse {
			clientError(w, http.StatusConflict, ErrCodeUserInUse,
				"user cannot be deleted as it is associated with previous orders") // 409
			return
		}
		if err != nil {
			contextLogger.Errorf("app: a.Service.DeleteUser(ctx, userID=%q) failed with error: %+v", userID, err)
			w.WriteHeader(http.StatusInternalServerError) // 500
			return
		}
		w.Header().Del("Content-Type")
		w.WriteHeader(http.StatusNoContent) // 204
	}
}
