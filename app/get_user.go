package app

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi"
	log "github.com/sirupsen/logrus"
)

// GetUserHandler returns an http.HandlerFunc that calls the service API
// to retrieve a list of users.
func (a *App) GetUserHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		contextLogger := log.WithContext(ctx)
		contextLogger.Info("App: GetUserHandler called")

		userID := chi.URLParam(r, "id")
		user, err := a.Service.GetUser(ctx, userID)
		if err != nil {
			contextLogger.Errorf("service GetUser(ctx, userID=%q) error: %+v", userID, err)
			w.WriteHeader(http.StatusInternalServerError) // 500 Internal Server Error
			return
		}
		w.WriteHeader(http.StatusOK) // 200 OK
		json.NewEncoder(w).Encode(*user)
	}
}
