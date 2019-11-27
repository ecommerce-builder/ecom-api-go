package app

import (
	"encoding/json"
	"net/http"

	service "bitbucket.org/andyfusniakteam/ecom-api-go/service/firebase"
	log "github.com/sirupsen/logrus"
)

// ListAddressesHandler get a list of addresses
func (a *App) ListAddressesHandler() http.HandlerFunc {
	type response struct {
		Object string             `json:"object"`
		Data   []*service.Address `json:"data"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		contextLogger := log.WithContext(ctx)
		contextLogger.Info("app: ListAddressesHandler started")

		userID := r.URL.Query().Get("user_id")
		if !IsValidUUID(userID) {
			// TODO: check admin privileges and return the whole list
			serverError(w, http.StatusBadRequest, ErrCodeBadRequest, "query parameter user_id must be a valid v4 UUID")
			return
		}
		contextLogger.Debugf("app: query param user_id=%q", userID)

		addresses, err := a.Service.GetAddresses(ctx, userID)
		if err == service.ErrUserNotFound {
			clientError(w, http.StatusNotFound, ErrCodeUserNotFound, "user not found") // 404
			return
		}
		if err != nil {
			contextLogger.Errorf("app: a.Service.GetAddresses(ctx, customerID=%q) error: %v", userID, err)
			w.WriteHeader(http.StatusInternalServerError) // 500 Internal Server Error
			return
		}

		list := response{
			Object: "list",
			Data:   addresses,
		}
		w.WriteHeader(http.StatusOK) // 200 OK
		json.NewEncoder(w).Encode(&list)
	}
}
