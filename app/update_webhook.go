package app

import (
	"encoding/json"
	"net/http"

	service "bitbucket.org/andyfusniakteam/ecom-api-go/service/firebase"
	"github.com/go-chi/chi"
	log "github.com/sirupsen/logrus"
)

type updateWebhookRequestBody struct {
	URL     *string        `json:"url,omitempty"`
	Events  *webhookEvents `json:"events,omitempty"`
	Enabled *bool          `json:"enabled,omitempty"`
}

// UpdateWebhookHandler returns a http.HandlerFunc that updates a webhook's
// url, events and enabled status.
func (a *App) UpdateWebhookHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		contextLogger := log.WithContext(ctx)
		contextLogger.Info("app: UpdateWebhookHandler started")

		// parse request body
		// example
		// {
		//   "url": "https://example.com/webhooks",
		//   "events": {
		//     "object": "list",
		//     "data": [ "order.created" ]
		//   }
		// }, enabled": true }
		var request updateWebhookRequestBody
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			clientError(w, http.StatusBadRequest, ErrCodeBadRequest, err.Error())
			return
		}

		ok, message := validateUpdateWebhookRequest(&request)
		if !ok {
			clientError(w, http.StatusBadRequest, ErrCodeBadRequest,
				message) // 400
			return
		}

		webhookID := chi.URLParam(r, "id")
		if !IsValidUUID(webhookID) {
			contextLogger.Warnf("400 Bad Request - invalid webhook %s", webhookID)
			clientError(w, http.StatusBadRequest, ErrCodeBadRequest,
				"path parameter id must be a valid v4 uuid") // 400
			return
		}

		var events []string
		if request.Events != nil {
			events = request.Events.Data
		}
		webhook, err := a.Service.UpdateWebhook(ctx, webhookID, request.URL, events, request.Enabled)
		if err == service.ErrWebhookNotFound {
			contextLogger.Warnf("404 Not Found - webhook %s not found", webhookID)
			clientError(w, http.StatusNotFound, ErrCodeWebhookNotFound,
				"webhook not found") // 404
			return
		}
		if err == service.ErrEventTypeNotFound {
			contextLogger.Warn("404 Not Found - one or more events not found")
			clientError(w, http.StatusNotFound, ErrCodeEventTypeNotFound,
				"one or more event types not recognised") // 404
			return
		}
		if err != nil {
			contextLogger.Errorf("app: a.Service.UpdateWebhook(ctx, webhookID=%q, url=%v, events=%v, enabled=%v) failed: %+v", webhookID, request.URL, events, request.Enabled, err)
			w.WriteHeader(http.StatusInternalServerError) // 500
			return
		}

		w.WriteHeader(http.StatusOK) // 200
		json.NewEncoder(w).Encode(&webhook)
	}
}

func validateUpdateWebhookRequest(request *updateWebhookRequestBody) (bool, string) {
	if request.URL == nil && request.Events == nil && request.Enabled == nil {
		return false, "you must set at least one attribute url, events or enabled"
	}

	return true, ""
}
