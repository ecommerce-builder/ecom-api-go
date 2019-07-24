package app

import (
	"net/http"

	log "github.com/sirupsen/logrus"
)

// StripeWebhookHandler returns a handler that processes Stripe webhook API calls.
func (a *App) StripeWebhookHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		contextLogger := log.WithContext(ctx)
		contextLogger.Info("App: StripeWebhookHandler started")

		err := a.Service.StripeProcessWebhook(ctx)
		if err != nil {
			contextLogger.Errorf("service StripeWebhook(ctx) error: %v", err)
			w.WriteHeader(http.StatusInternalServerError) // 500 Internal Server Error
			return
		}
		w.WriteHeader(http.StatusNoContent) // 204 No Content
	}
}
