package app

import (
	"net/http"

	service "bitbucket.org/andyfusniakteam/ecom-api-go/service/firebase"
	"github.com/go-chi/chi"
	log "github.com/sirupsen/logrus"
)

// DeleteWebhookHandler creates a handler function that deletes
// a webhook by id.
func (a *App) DeleteWebhookHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		contextLogger := log.WithContext(ctx)
		contextLogger.Info("app: DeleteWebhookHandler started")

		webhookID := chi.URLParam(r, "id")
		err := a.Service.DeleteWebhook(ctx, webhookID)
		if err == service.ErrWebhookNotFound {
			clientError(w, http.StatusNotFound, ErrCodeWebhookNotFound,
				"webhook not found") // 404
			return
		}
		if err != nil {
			contextLogger.Errorf("app: a.Service.DeleteWebhook(ctx, webhookID=%q) error: %+v", webhookID, err)
			w.WriteHeader(http.StatusInternalServerError) //
			return
		}
		w.Header().Del("Content-Type")
		w.WriteHeader(http.StatusNoContent) // 204
	}
}
