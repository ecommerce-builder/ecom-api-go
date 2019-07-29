package app

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	log "github.com/sirupsen/logrus"
	"github.com/stripe/stripe-go"
	"github.com/stripe/stripe-go/webhook"
)

// StripeWebhookHandler returns a handler that processes Stripe webhook API calls.
func (a *App) StripeWebhookHandler(stripeSigningSecret string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		contextLogger := log.WithContext(ctx)
		contextLogger.Info("App: StripeWebhookHandler started")

		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			contextLogger.Errorf("failed to read request body: %v\n", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		// Pass the request body & Stripe-Signature header
		// to ConstructEvent, along with the webhook signing key
		event, err := webhook.ConstructEvent(body, r.Header.Get("Stripe-Signature"), stripeSigningSecret)
		if err != nil {
			contextLogger.Errorf("failed to verify webhook signature: %v", err)
			w.WriteHeader(http.StatusBadRequest) // Return a 400 error on a bad signature
			return
		}

		// Handle the checkout.session.completed event
		if event.Type == "checkout.session.completed" {
			var session stripe.CheckoutSession
			err := json.Unmarshal(event.Data.Raw, &session)
			if err != nil {
				contextLogger.Errorf("failed to parse webhook JSON: %v", err)
				w.WriteHeader(http.StatusBadRequest)
				return
			}

			// Fulfill the purchase...
			err = a.Service.StripeProcessWebhook(ctx, session, body)
			if err != nil {
				contextLogger.Errorf("service StripeProcessWebhook(ctx, session=%v) error: %v", err, session)
				w.WriteHeader(http.StatusInternalServerError) // 500 Internal Server Error
				return
			}
		}

		w.WriteHeader(http.StatusNoContent) // 204 No Content
	}
}
