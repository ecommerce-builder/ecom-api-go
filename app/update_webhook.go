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

func validateUpdateWebhookRequest(request *updateWebhookRequestBody) (bool, string) {
	if request.URL == nil && request.Events == nil && request.Enabled == nil {
		return false, "you must set at least one attribute url, events or enabled"
	}

	return true, ""
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
			clientError(w, http.StatusBadRequest, ErrCodeBadRequest, message)
			return
		}

		webhookID := chi.URLParam(r, "id")
		var events []string
		if request.Events != nil {
			events = request.Events.Data
		}
		webhook, err := a.Service.UpdateWebhook(ctx, webhookID, request.URL, events, request.Enabled)
		if err != nil {
			if err == service.ErrWebhookNotFound {
				clientError(w, http.StatusNotFound, ErrCodeWebhookNotFound, "webhook not found")
				return
			} else if err == service.ErrEventTypeNotFound {
				clientError(w, http.StatusNotFound, ErrCodeEventTypeNotFound, "one or more event types not recognised")
				return
			}
			contextLogger.Errorf("app: a.Service.UpdateWebhook(ctx, webhookID=%q, url=%v, events=%v, enabled=%v) failed: %+v", webhookID, request.URL, events, request.Enabled, err)
			w.WriteHeader(http.StatusInternalServerError) // 500 Internal Server Error
			return
		}

		w.WriteHeader(http.StatusOK) // 200 OK
		json.NewEncoder(w).Encode(&webhook)
	}
}
