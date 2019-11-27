package app

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi"
	log "github.com/sirupsen/logrus"
)

// StripeCheckoutHandler returns a handler that returns a Stripe Checkout Session ID.
func (a *App) StripeCheckoutHandler(stripeSuccessURL, stripeCancelURL string) http.HandlerFunc {
	type stripeCheckoutResponseBody struct {
		Object            string `json:"object"`
		CheckoutSessionID string `json:"checkout_session_id"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		contextLogger := log.WithContext(ctx)
		contextLogger.Info("app: StripeCheckoutHandler started")

		orderID := chi.URLParam(r, "id")
		if !IsValidUUID(orderID) {
			contextLogger.Warn("app: path param is not a valid v4 uuid")
			clientError(w, http.StatusBadRequest, ErrCodeBadRequest,
				"path parameter id must be set to a valid v4 uuid")
			return
		}

		contextLogger.Debugf("app: order id %s", orderID)
		sid, err := a.Service.StripeCheckout(ctx, orderID, stripeSuccessURL, stripeCancelURL)
		if err != nil {
			contextLogger.Errorf("app: StripeCheckout(ctx, %q) error: %v",
				orderID, err)
			w.WriteHeader(http.StatusInternalServerError) // 500
			return
		}
		res := stripeCheckoutResponseBody{
			Object:            "stripe_checkout_session",
			CheckoutSessionID: sid,
		}
		w.WriteHeader(http.StatusCreated) // 201
		json.NewEncoder(w).Encode(res)
	}
}
