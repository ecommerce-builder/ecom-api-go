package app

import (
	"encoding/json"
	"net/http"

	service "bitbucket.org/andyfusniakteam/ecom-api-go/service/firebase"
	"github.com/go-chi/chi"
	log "github.com/sirupsen/logrus"
)

// GetUserHandler returns an http.HandlerFunc that calls the service API
// to retrieve a list of users.
func (a *App) GetUserHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		contextLogger := log.WithContext(ctx)
		contextLogger.Info("app: GetUserHandler called")

		userID := chi.URLParam(r, "id")
		user, err := a.Service.GetUser(ctx, userID)
		if err == service.ErrUserNotFound {
			clientError(w, http.StatusNotFound, ErrCodeUserNotFound, "user not found") // 404
			return
		}
		if err != nil {
			contextLogger.Errorf("app: GetUser(ctx, userID=%q) error: %+v", userID, err)
			w.WriteHeader(http.StatusInternalServerError) // 500 Internal Server Error
			return
		}
		w.WriteHeader(http.StatusOK) // 200 OK
		json.NewEncoder(w).Encode(*user)
	}
}
