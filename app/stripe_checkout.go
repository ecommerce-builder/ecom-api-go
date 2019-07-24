package app

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi"
	log "github.com/sirupsen/logrus"
)

// StripeCheckoutHandler returns a handler that returns a Stripe Checkout Session ID.
func (a *App) StripeCheckoutHandler() http.HandlerFunc {
	type stripeCheckoutResponseBody struct {
		Object            string `json:"object"`
		CheckoutSessionID string `json:"checkout_session_id"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		contextLogger := log.WithContext(ctx)
		contextLogger.Info("App: StripeCheckoutHandler started")

		id := chi.URLParam(r, "id")
		sid, err := a.Service.StripeCheckout(ctx, id)
		if err != nil {
			contextLogger.Errorf("service StripeCheckout(ctx, %q) error: %v", id, err)
			w.WriteHeader(http.StatusInternalServerError) // 500 Internal Server Error
			return
		}
		res := stripeCheckoutResponseBody{
			Object:            "stripe_checkout_session",
			CheckoutSessionID: sid,
		}
		w.WriteHeader(http.StatusCreated) // 201 Created
		json.NewEncoder(w).Encode(res)
	}
}
