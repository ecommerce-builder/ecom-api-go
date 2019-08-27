package app

import (
	"encoding/json"
	"net/http"

	service "bitbucket.org/andyfusniakteam/ecom-api-go/service/firebase"
	"github.com/go-chi/chi"
	log "github.com/sirupsen/logrus"
)

// ListUsersDevKeysHandler get a list of addresses
func (a *App) ListUsersDevKeysHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		contextLogger := log.WithContext(ctx)
		contextLogger.Info("App: ListUsersDevKeysHandler started")

		userID := chi.URLParam(r, "id")
		apiKeys, err := a.Service.ListUsersDevKeys(ctx, userID)
		if err != nil {
			if err == service.ErrUserNotFound {
				w.WriteHeader(http.StatusNotFound) // 404 Not Found
				return
			}
			contextLogger.Errorf("service ListUsersDevKeys(ctx, userID=%q) error: %v", userID, err)
			w.WriteHeader(http.StatusInternalServerError) // 500 Internal Server Error
			return
		}
		w.WriteHeader(http.StatusOK) // 200 OK
		json.NewEncoder(w).Encode(apiKeys)
	}
}
