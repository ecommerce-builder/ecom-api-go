package app

import (
	"encoding/json"
	"net/http"

	service "bitbucket.org/andyfusniakteam/ecom-api-go/service/firebase"
	log "github.com/sirupsen/logrus"
)

// GenerateUserDevKeyHandler creates a new API Key for a given user
func (a *App) GenerateUserDevKeyHandler() http.HandlerFunc {
	type requestBody struct {
		UserID string `json:"user_id"`
	}

	type responseBody struct {
		Object string `json:"object"`
		*service.UserDevKey
	}

	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		contextLogger := log.WithContext(ctx)
		contextLogger.Info("App: GenerateUserDevKeyHandler started")

		request := requestBody{}
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(struct {
				Status  int    `json:"status"`
				Code    string `json:"code"`
				Message string `json:"message"`
			}{
				http.StatusBadRequest,
				ErrCodeBadRequest,
				"bad request",
			})
			return
		}

		cdk, err := a.Service.GenerateUserDevKey(ctx, request.UserID)
		if err != nil {
			if err == service.ErrUserNotFound {
				w.WriteHeader(http.StatusNotFound) // Not Found
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
			contextLogger.Errorf("service GenerateUserAPIKey(ctx, userID=%q) error: %v", request.UserID, err)
			w.WriteHeader(http.StatusInternalServerError) // 500 Internal Server Error
			return
		}
		response := responseBody{
			Object:     "developer_key",
			UserDevKey: cdk,
		}
		w.WriteHeader(http.StatusCreated) // 201 Created
		json.NewEncoder(w).Encode(&response)
	}
}
