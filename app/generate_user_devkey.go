package app

import (
	"encoding/json"
	"net/http"

	service "bitbucket.org/andyfusniakteam/ecom-api-go/service/firebase"
	log "github.com/sirupsen/logrus"
)

type generateUserDevKeyRequestBody struct {
	UserID *string `json:"user_id"`
}

func validateGenerateUserDevKeyRequestBody(request *generateUserDevKeyRequestBody) (bool, string) {
	// user_id attribute
	userID := request.UserID
	if userID == nil {
		return false, "user_id attribute must be set"
	}
	if !IsValidUUID(*userID) {
		return false, "user_id attribute must be a valid v4 UUID"
	}
	return true, ""
}

// GenerateUserDevKeyHandler creates a new API Key for a given user
func (a *App) GenerateUserDevKeyHandler() http.HandlerFunc {
	type responseBody struct {
		Object string `json:"object"`
		*service.DeveloperKey
	}

	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		contextLogger := log.WithContext(ctx)
		contextLogger.Info("app: GenerateUserDevKeyHandler started")

		request := generateUserDevKeyRequestBody{}
		dec := json.NewDecoder(r.Body)
		dec.DisallowUnknownFields()
		if err := dec.Decode(&request); err != nil {
			clientError(w, http.StatusBadRequest, ErrCodeBadRequest, err.Error())
			return
		}

		ok, message := validateGenerateUserDevKeyRequestBody(&request)
		if !ok {
			clientError(w, http.StatusBadRequest, ErrCodeBadRequest, message)
			return
		}

		developerKey, err := a.Service.GenerateUserDevKey(ctx, *request.UserID)
		if err != nil {
			if err == service.ErrUserNotFound {
				// 404 Not Found
				clientError(w, http.StatusNotFound, ErrCodeUserNotFound, "user not found")
				return
			}
			contextLogger.Errorf("app: GenerateUserAPIKey(ctx, userID=%q) error: %v", request.UserID, err)
			w.WriteHeader(http.StatusInternalServerError) // 500 Internal Server Error
			return
		}
		w.WriteHeader(http.StatusCreated) // 201 Created
		json.NewEncoder(w).Encode(&developerKey)
	}
}
