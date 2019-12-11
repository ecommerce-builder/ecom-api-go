package firebase

import (
	"context"
	"fmt"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/stripe/stripe-go"
	"github.com/stripe/stripe-go/checkout/session"
)

// StripeCheckout generates a stripe checkout session
func (s *Service) StripeCheckout(ctx context.Context, orderID, stripeSuccessURL, stripeCancelURL string) (string, error) {
	contextLogger := log.WithContext(ctx)
	contextLogger.Debugf("service: StripeCheckout(ctx, orderID=%q, stripeSuccessURL=%q, stripeCancelURL=%q)",
		orderID, stripeSuccessURL, stripeCancelURL)

	order, err := s.GetOrder(ctx, orderID)
	if err != nil {
		// TODO: deal with ErrOrderNotFound and ErrOrderItemsNotFound
		return "", errors.Wrapf(err, "s.GetOrder(ctx, orderID=%s", orderID)
	}
	fmt.Println(order)

	items := make([]*stripe.CheckoutSessionLineItemParams, 0, len(order.Items))

	for _, i := range order.Items {
		var vatMultiplier float64
		if i.TaxCode == "T20" {
			vatMultiplier = 1.2
		} else {
			vatMultiplier = 1.0
		}
		desc := fmt.Sprintf("%d x %s", i.Qty, i.Name)

		stripeUnitPrice := int64((float64(i.UnitPrice) * vatMultiplier) / 100.0)
		t := stripe.CheckoutSessionLineItemParams{
			Name:        stripe.String(i.SKU),
			Description: stripe.String(desc),
			Amount:      stripe.Int64(stripeUnitPrice),
			Currency:    stripe.String(string(stripe.CurrencyGBP)),
			Quantity:    stripe.Int64(int64(i.Qty)),
		}
		items = append(items, &t)

		contextLogger.Infof("service: stripe line item added - product id=%s, path=%s, sku=%s, name=%q, qty=%d, unitPrice=%v, currency=%s, discount=%d, taxCode=%s, VAT=%d", i.ID, i.Path, i.SKU, i.Name, i.Qty, i.UnitPrice, i.Currency, i.Discount, i.TaxCode, i.VAT)
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
		SuccessURL: stripe.String(stripeSuccessURL),
		CancelURL:  stripe.String(stripeCancelURL),
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
func (s *Service) StripeProcessWebhook(ctx context.Context, session stripe.CheckoutSession, body []byte) (*Order, error) {
	contextLogger := log.WithContext(ctx)
	contextLogger.Debugf("service: session.ClientReferenceID=%q", session.ClientReferenceID)
	contextLogger.Debugf("service: session.PaymentIntentID=%q", session.PaymentIntent.ID)

	orow, oirows, bill, ship, err := s.model.RecordPayment(ctx,
		session.ClientReferenceID, session.PaymentIntent.ID, body)
	if err != nil {
		return nil, errors.Wrapf(err,
			"s.model.RecordPayment(ctx, orderID=%s, pi=%s",
			session.ClientReferenceID,
			session.PaymentIntent.ID)
	}

	orderItems := make([]*OrderItem, 0, len(oirows))
	for _, row := range oirows {
		oi := OrderItem{
			Object:    "order_item",
			ID:        row.UUID,
			Path:      row.Path,
			SKU:       row.SKU,
			Name:      row.Name,
			Qty:       row.Qty,
			UnitPrice: row.UnitPrice,
			Currency:  row.Currency,
			Discount:  row.Discount,
			TaxCode:   row.TaxCode,
			VAT:       row.VAT,
			Created:   &row.Created,
		}
		orderItems = append(orderItems, &oi)
	}

	order := Order{
		Object:  "order",
		ID:      orow.UUID,
		OrderID: orow.ID,
		Status:  orow.Status,
		Payment: orow.Payment,
		User: &OrderUser{
			ID:          orow.UsrUUID,
			ContactName: orow.ContactName,
			Email:       orow.Email,
		},
		Billing: &OrderAddress{
			ContactName: bill.ContactName,
			Addr1:       bill.Addr1,
			Addr2:       bill.Addr2,
			City:        bill.City,
			County:      bill.County,
			Postcode:    bill.Postcode,
			Country:     bill.CountryCode,
		},
		Shipping: &OrderAddress{
			ContactName: ship.ContactName,
			Addr1:       ship.Addr1,
			Addr2:       ship.Addr2,
			City:        ship.City,
			County:      ship.County,
			Postcode:    ship.Postcode,
			Country:     ship.CountryCode,
		},
		Currency:    orow.Currency,
		TotalExVAT:  orow.TotalExVAT,
		VATTotal:    orow.VATTotal,
		TotalIncVAT: orow.TotalIncVAT,
		Items:       orderItems,
		Created:     orow.Created,
		Modified:    orow.Modified,
	}
	if err := s.PublishTopicEvent(ctx, EventOrderUpdated, &order); err != nil {
		return nil, errors.Wrapf(err,
			"service: s.PublishTopicEvent(ctx, event=%q, data=%v) failed",
			EventOrderUpdated, order)
	}
	contextLogger.Infof("service: EventOrderUpdated published")
	return &order, nil
}
