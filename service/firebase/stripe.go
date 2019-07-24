package firebase

import (
	"context"
	"fmt"
	"math"

	"github.com/pkg/errors"
	"github.com/stripe/stripe-go"
	"github.com/stripe/stripe-go/checkout/session"
)

// StripeCheckout generates a stripe checkout session
func (s *Service) StripeCheckout(ctx context.Context, orderID string) (string, error) {
	order, err := s.GetOrder(ctx, orderID)
	if err != nil {
		return "", errors.Wrapf(err, "s.GetOrder(ctx, orderID=%s", orderID)
	}
	fmt.Printf("%#v\n", order)

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

	params := &stripe.CheckoutSessionParams{
		//CustomerEmail: stripe.String(email),
		//Customer:          stripe.String("cus_Amzea4PdYUwsqa"),
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
	return cs.ID, nil
}

// StripeProcessWebhook processes the webhook called by the Stripe system.
func (s *Service) StripeProcessWebhook(ctx context.Context) error {
	return nil
}
