package app

import (
	"crypto/subtle"
	"encoding/base64"
	"encoding/json"
	"net/http"

	service "bitbucket.org/andyfusniakteam/ecom-api-go/service/firebase"
	log "github.com/sirupsen/logrus"
)

// PubSubBroadcastHandler listens for Cloud Pub/Sub callbacks from
// push requests triggered by previous published events.
// For each received callback the Webhook URL will be called to
// deliver the payload.
func (a *App) PubSubBroadcastHandler(secret string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		contextLogger := log.WithContext(ctx)
		contextLogger.Info("app: PubSubBroadcastHandler started")

		token := r.URL.Query().Get("token")
		if subtle.ConstantTimeCompare([]byte(token), []byte(secret)) == 0 {
			log.Warn("app: broadcast push endpoint called with invalid token query parameter")
			unauthorized(w)
			return
		}
		log.Infof("app: pubsub broadcast push endpoint auth succeeded")

		var req service.PubSubPayload
		err := json.NewDecoder(r.Body).Decode(&req)
		if err != nil {
			log.Errorf("app: failed to decode pubsub message")
			clientError(w, http.StatusBadRequest, ErrCodeBadRequest, err.Error())
			return
		}
		defer r.Body.Close()

		decoded, err := base64.StdEncoding.DecodeString(req.Message.Data)
		if err != nil {
			log.Errorf("app: base64.StdEncoding.DecodeString(data=%v) failed: %+v", req.Message.Data, err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		v, err := service.DecodeEventData(req.Message.Attributes.Event, decoded)
		if err != nil {
			log.Errorf("app: service.DecodeEventData(event=%q, data=%v) failed: %+v", req.Message.Attributes.Event, decoded, err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		err = a.Service.CallWebhook(ctx, req.Message.MessageID, req.Message.Attributes.WebhookID, req.Message.Attributes.Event, v)
		if err == service.ErrWebhookPostFailed {
			clientError(w, http.StatusConflict, ErrCodeWebhookPostFailed, err.Error())
			return
		}
		if err == service.ErrWebhookNotFound {
			clientError(w, http.StatusNotFound, ErrCodeWebhookNotFound, "pubsub webhook not found")
			return
		}
		if err != nil {
			log.Errorf("app: a.Service.CallWebhook(ctx, webhookID=%q, eventName=%q, data=%s) failed: %+v", req.Message.Attributes.WebhookID, req.Message.Attributes.Event, req.Message.Data, err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		log.Infof("app: call to webhook id=%q succeeded", req.Message.Attributes.WebhookID)
		w.WriteHeader(http.StatusOK)
	}
}
