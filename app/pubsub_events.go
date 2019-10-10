package app

import (
	"crypto/subtle"
	"encoding/json"
	"net/http"

	service "bitbucket.org/andyfusniakteam/ecom-api-go/service/firebase"
	log "github.com/sirupsen/logrus"
)

// PubSubEventHandler listens for Cloud Pub/Sub callbacks from
// push requests triggered by previous published events.
// For each webhook registered in the system, PubSubEventHandler
// will publish an individual pubsub event on a different topic.
func (a *App) PubSubEventHandler(secret string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		contextLogger := log.WithContext(ctx)
		contextLogger.Info("app: PubSubEventHandler started")

		token := r.URL.Query().Get("token")
		if subtle.ConstantTimeCompare([]byte(token), []byte(secret)) == 0 {
			log.Warn("app: event push endpoint called with invalid token query parameter")
			unauthorized(w)
			return
		}
		log.Infof("app: pubsub event push endpoint auth succeeded")

		var req service.PubSubPayload
		err := json.NewDecoder(r.Body).Decode(&req)
		if err != nil {
			log.Errorf("app: failed to decode pubsub message")
			clientError(w, http.StatusBadRequest, ErrCodeBadRequest, err.Error())
			return
		}
		defer r.Body.Close()

		err = a.Service.BroadcastEvents(ctx, req.Message)
		if err != nil {
			log.Errorf("app: a.Service.BroadcastEvents(ctx, msg=%v)", req.Message)
			w.WriteHeader(http.StatusInternalServerError) // 500 Internal Server Error
			return
		}
		w.WriteHeader(http.StatusOK)
	}
}
