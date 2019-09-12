package app

import (
	"encoding/json"
	"net/http"

	service "bitbucket.org/andyfusniakteam/ecom-api-go/service/firebase"
	log "github.com/sirupsen/logrus"
)

func validateListUsersDevKeysRequestBody(request *generateUserDevKeyRequestBody) (bool, string) {
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

// ListUsersDevKeysHandler get a list of addresses
func (a *App) ListUsersDevKeysHandler() http.HandlerFunc {
	type listResponse struct {
		Object string                  `json:"object"`
		Data   []*service.DeveloperKey `json:"data"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		contextLogger := log.WithContext(ctx)
		contextLogger.Info("app: ListUsersDevKeysHandler started")

		userID := r.URL.Query().Get("user_id")
		if userID == "" {
			clientError(w, http.StatusBadRequest, ErrCodeBadRequest, "user_id query param must be set")
			return
		}
		if !IsValidUUID(userID) {
			clientError(w, http.StatusBadRequest, ErrCodeBadRequest, "user_id query param must be a valid v4 UUID")
			return
		}

		developerKeys, err := a.Service.ListUsersDevKeys(ctx, userID)
		if err != nil {
			if err == service.ErrUserNotFound {
				clientError(w, http.StatusNotFound, ErrCodeUserNotFound, "user not found")
				return
			}
			contextLogger.Errorf("app: ListUsersDevKeys(ctx, userID=%q) error: %+v", userID, err)
			w.WriteHeader(http.StatusInternalServerError) // 500 Internal Server Error
			return
		}

		list := listResponse{
			Object: "list",
			Data:   developerKeys,
		}

		w.WriteHeader(http.StatusOK) // 200 OK
		json.NewEncoder(w).Encode(&list)
	}
}
