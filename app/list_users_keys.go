package app

import (
	"encoding/json"
	"net/http"

	service "bitbucket.org/andyfusniakteam/ecom-api-go/service/firebase"
	log "github.com/sirupsen/logrus"
)

// ListUsersDevKeysHandler get a list of addresses
func (a *App) ListUsersDevKeysHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		contextLogger := log.WithContext(ctx)
		contextLogger.Info("App: ListUsersDevKeysHandler started")

		userID := r.URL.Query().Get("user_id")
		developerKeys, err := a.Service.ListUsersDevKeys(ctx, userID)
		if err != nil {
			if err == service.ErrUserNotFound {
				w.WriteHeader(http.StatusNotFound) // 404 Not Found
				json.NewEncoder(w).Encode(struct {
					Status  int    `json:"status"`
					Code    string `json:"code"`
					Message string `json:"message"`
				}{
					http.StatusNotFound,
					ErrCodeUserNotFound,
					"user not found",
				})
				return

			}
			contextLogger.Errorf("service ListUsersDevKeys(ctx, userID=%q) error: %v", userID, err)
			w.WriteHeader(http.StatusInternalServerError) // 500 Internal Server Error
			return
		}
		w.WriteHeader(http.StatusOK) // 200 OK
		json.NewEncoder(w).Encode(developerKeys)
	}
}
