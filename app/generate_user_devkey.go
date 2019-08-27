package app

import (
	"encoding/json"
	"net/http"

	service "bitbucket.org/andyfusniakteam/ecom-api-go/service/firebase"
	"github.com/go-chi/chi"
	log "github.com/sirupsen/logrus"
)

// GenerateUserDevKeyHandler creates a new API Key for a given user
func (a *App) GenerateUserDevKeyHandler() http.HandlerFunc {
	type userDevKeyResponseBody struct {
		Object string `json:"object"`
		*service.UserDevKey
	}

	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		contextLogger := log.WithContext(ctx)
		contextLogger.Info("App: GenerateUserDevKeyHandler started")

		userID := chi.URLParam(r, "id")
		cdk, err := a.Service.GenerateUserDevKey(ctx, userID)
		if err != nil {
			if err == service.ErrUserNotFound {
				w.WriteHeader(http.StatusNotFound) // 404 Not Found
				return
			}
			contextLogger.Errorf("service GenerateUserAPIKey(ctx, userID=%q) error: %v", userID, err)
			w.WriteHeader(http.StatusInternalServerError) // 500 Internal Server Error
			return
		}
		res := userDevKeyResponseBody{
			Object:     "developer_key",
			UserDevKey: cdk,
		}
		w.WriteHeader(http.StatusCreated) // 201 Created
		json.NewEncoder(w).Encode(res)
	}
}
