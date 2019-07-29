package firebase

import (
	"context"
	"math"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/stripe/stripe-go"
	"github.com/stripe/stripe-go/checkout/session"
)

// StripeCheckout generates a stripe checkout session
func (s *Service) StripeCheckout(ctx context.Context, orderID string) (string, error) {
	order, err := s.GetOrder(ctx, orderID)
	if err != nil {
		// TODO: deal with ErrOrderNotFound and ErrOrderItemsNotFound
		return "", errors.Wrapf(err, "s.GetOrder(ctx, orderID=%s", orderID)
	}

	items := make([]*stripe.CheckoutSessionLineItemParams, 0, len(order.Items))

	for _, i := range order.Items {
		t := stripe.CheckoutSessionLineItemParams{
			Name:        stripe.String(i.SKU),
			Description: stripe.String(i.Name),
			Amount:      stripe.Int64(int64(i.UnitPrice) + int64(math.Round(float64(i.UnitPrice)*0.2))),
			Currency:    stripe.String(string(stripe.CurrencyGBP)),
			Quantity:    stripe.Int64(int64(i.Qty)),
		}
		items = append(items, &t)
	}

	paymentIntentDataParams := &stripe.CheckoutSessionPaymentIntentDataParams{}
	paymentIntentDataParams.AddMetadata("order_id", orderID)

	params := &stripe.CheckoutSessionParams{
		//CustomerEmail: stripe.String(email),
		//Customer:          stripe.String("cus_Amzea4PdYUwsqa"),
		PaymentIntentData: paymentIntentDataParams,
		ClientReferenceID: stripe.String(orderID),
		PaymentMethodTypes: stripe.StringSlice([]string{
			"card",
		}),
		LineItems:  items,
		SuccessURL: stripe.String("https://example.com/success"),
		CancelURL:  stripe.String("https://example.com/cancel"),
	}

	cs, err := session.New(params)
	if err != nil {
		return "", errors.Wrap(err, "failed to create new Stripe session")
	}
	log.Debugf("stripe success URL %s", cs.SuccessURL)
	log.Debugf("stripe cancel URL %s", cs.CancelURL)
	log.Debugf("stripe checkout session id %s", cs.ID)
	log.Debugf("stripe payment intent id %s", cs.PaymentIntent.ID)

	err = s.SetStripePaymentIntentID(ctx, orderID, cs.PaymentIntent.ID)
	if err != nil {
		return "", errors.Wrapf(err, "s.SetStripePaymentIntentID(ctx, orderID=%s, pi=%s", orderID, cs.PaymentIntent.ID)
	}
	return cs.ID, nil
}

// StripeProcessWebhook processes the webhook called by the Stripe system.
func (s *Service) StripeProcessWebhook(ctx context.Context, session stripe.CheckoutSession, body []byte) error {
	contextLogger := log.WithContext(ctx)
	contextLogger.Debugf("StripeProcessWebhook: session.ClientReferenceID %s", session.ClientReferenceID)
	contextLogger.Debugf("StripeProcessWebhook: session.PaymentIntentID %s", session.PaymentIntent.ID)

	err := s.model.RecordPayment(ctx, session.ClientReferenceID, session.PaymentIntent.ID, body)
	if err != nil {
		return errors.Wrapf(err, "s.model.RecordPayment(ctx, orderID=%s, pi=%s", session.ClientReferenceID, session.PaymentIntent.ID)
	}
	return nil
}
