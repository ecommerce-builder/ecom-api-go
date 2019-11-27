package app

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi"

	service "bitbucket.org/andyfusniakteam/ecom-api-go/service/firebase"
	log "github.com/sirupsen/logrus"
)

// GetWebhookHandler creates a handler function that returns a
// single webhook.
func (a *App) GetWebhookHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		contextLogger := log.WithContext(ctx)
		contextLogger.Info("app: GetWebhookHandler started")

		webhookID := chi.URLParam(r, "id")
		if !IsValidUUID(webhookID) {
			clientError(w, http.StatusBadRequest, ErrCodeBadRequest, "url parameter must be a valid v4 UUID")
			return
		}

		webhook, err := a.Service.GetWebhook(ctx, webhookID)
		if err == service.ErrWebhookNotFound {
			clientError(w, http.StatusNotFound, ErrCodeWebhookNotFound, "webhook not found")
			return
		}
		if err != nil {
			contextLogger.Errorf("app: a.Service.GetWebhook(ctx, webhookID=%q) failed: %+v", webhookID, err)
			w.WriteHeader(http.StatusInternalServerError) // 500
			return
		}
		w.WriteHeader(http.StatusOK) // 200 OK
		json.NewEncoder(w).Encode(&webhook)
	}
}
