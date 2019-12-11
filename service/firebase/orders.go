package firebase

import (
	"context"
	"time"

	"bitbucket.org/andyfusniakteam/ecom-api-go/model/postgres"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

// ErrCartEmpty is returned whilst attempting to place an order with an
// empty cart.
var ErrCartEmpty = errors.New("service: failed to place an order with empty cart")

// ErrOrderNotFound error.
var ErrOrderNotFound = errors.New("service: order not found")

// ErrOrderItemsNotFound error.
var ErrOrderItemsNotFound = errors.New("service: order items not found")

// NewOrderAddressRequest contains the new address request body
type NewOrderAddressRequest struct {
	ContactName *string `json:"contact_name"`
	Addr1       *string `json:"addr1"`
	Addr2       *string `json:"addr2"`
	City        *string `json:"city"`
	County      *string `json:"county"`
	Postcode    *string `json:"postcode"`
	CountryCode *string `json:"country_code"`
}

// OrderAddress contains details of an address within an Order.
type OrderAddress struct {
	ContactName string  `json:"contact_name"`
	Addr1       string  `json:"addr1"`
	Addr2       *string `json:"addr2,omitempty"`
	City        string  `json:"city"`
	County      *string `json:"county,omitempty"`
	Postcode    string  `json:"postcode"`
	Country     string  `json:"country_code"`
}

// OrderItem contains details of a line item within an Order.
type OrderItem struct {
	Object    string     `json:"object"`
	ID        string     `json:"id"`
	Path      string     `json:"path"`
	SKU       string     `json:"sku"`
	Name      string     `json:"name"`
	Qty       int        `json:"qty"`
	UnitPrice int        `json:"unit_price"`
	Currency  string     `json:"currency"`
	Discount  *int       `json:"discount,omitempty"`
	TaxCode   string     `json:"tax_code"`
	VAT       int        `json:"vat"`
	Created   *time.Time `json:"created,omitempty"`
}

// OrderUser contains details of the guest or user that placed the order.
type OrderUser struct {
	ID          *string `json:"id,omitempty"`
	ContactName *string `json:"contact_name,omitempty"`
	Email       *string `json:"email,omitempty"`
}

// Order contains details of an existing order.
type Order struct {
	Object      string        `json:"object"`
	ID          string        `json:"id"`
	OrderID     int           `json:"order_id"`
	Status      string        `json:"status"`
	Payment     string        `json:"payment"`
	User        *OrderUser    `json:"user"`
	Billing     *OrderAddress `json:"billing_address"`
	Shipping    *OrderAddress `json:"shipping_address"`
	Currency    string        `json:"currency"`
	TotalExVAT  int           `json:"total_ex_vat"`
	VATTotal    int           `json:"vat_total"`
	TotalIncVAT int           `json:"total_inc_vat"`
	Items       []*OrderItem  `json:"items"`
	Created     time.Time     `json:"created"`
	Modified    time.Time     `json:"modified"`
}

// PlaceGuestOrder places a new guest order.
func (s *Service) PlaceGuestOrder(ctx context.Context, cartID, contactName,
	email string, billing, shipping *NewOrderAddressRequest) (*Order, error) {
	contextLogger := log.WithContext(ctx)
	contextLogger.Debugf("service: PlaceGuestOrder(ctx, cartID=%q, contactName=%s, email=%s, ...)",
		cartID, contactName, email)

	pgBilling := postgres.NewOrderAddress{
		ContactName: *billing.ContactName,
		Addr1:       *billing.Addr1,
		Addr2:       billing.Addr2,
		City:        *billing.City,
		County:      billing.County,
		Postcode:    *billing.Postcode,
		CountryCode: *billing.CountryCode,
	}
	pgShipping := postgres.NewOrderAddress{
		ContactName: *shipping.ContactName,
		Addr1:       *shipping.Addr1,
		Addr2:       shipping.Addr2,
		City:        *shipping.City,
		County:      shipping.County,
		Postcode:    *shipping.Postcode,
		CountryCode: *shipping.CountryCode,
	}

	orow, oirows, bill, ship, err := s.model.AddGuestOrder(ctx,
		cartID, contactName, email, &pgBilling, &pgShipping)
	if err == postgres.ErrCartNotFound {
		return nil, ErrCartNotFound
	}
	if err == postgres.ErrCartEmpty {
		return nil, ErrCartEmpty
	}
	if err != nil {
		return nil, errors.Wrap(err, "service: s.model.AddGuestOrder(ctx, ...)")

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
	if err := s.PublishTopicEvent(ctx, EventOrderCreated, &order); err != nil {
		return nil, errors.Wrapf(err,
			"service: s.PublishTopicEvent(ctx, event=%q, data=%v) failed",
			EventOrderCreated, order)
	}
	contextLogger.Infof("service: EventOrderCreated published")
	return &order, nil
}

// PlaceOrder places a new order in the system for an existing user.
func (s *Service) PlaceOrder(ctx context.Context, cartID, userID, billingID, shippingID string) (*Order, error) {
	contextLogger := log.WithContext(ctx)
	contextLogger.Debugf("service: PlaceOrder(ctx, cartID=%q, customerID=%q, billingID=%q, shippingID=%q)",
		cartID, userID, billingID, shippingID)

	orow, oirows, urow, bill, ship, err := s.model.AddOrder(ctx, cartID, userID, billingID, shippingID)
	if err == postgres.ErrCartNotFound {
		return nil, ErrCartNotFound
	}
	if err == postgres.ErrCartEmpty {
		return nil, ErrCartEmpty
	}
	if err == postgres.ErrAddressNotFound {
		return nil, ErrAddressNotFound
	}
	if err != nil {
		return nil, errors.Wrap(err, "service: s.model.AddOrder(ctx, ...) failed")
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
			ID: &urow.UUID,
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
	if err := s.PublishTopicEvent(ctx, EventOrderCreated, &order); err != nil {
		return nil, errors.Wrapf(err,
			"service: s.PublishTopicEvent(ctx, event=%q, data=%v) failed",
			EventOrderCreated, order)
	}
	contextLogger.Infof("service: EventOrderCreated published")
	return &order, nil
}

// GetOrders returns a list of order summaries.
func (s *Service) GetOrders(ctx context.Context) ([]*Order, error) {
	contextLogger := log.WithContext(ctx)
	contextLogger.Debug("service: GetOrders(ctx)")

	rows, err := s.model.GetOrders(ctx)
	if err != nil {
		return nil, errors.Wrapf(err, "service: s.model.GetOrders(ctx)")
	}
	contextLogger.Debugf("service: s.model.GetOrders(ctx) returned %d rows", len(rows))

	orders := make([]*Order, 0, len(rows))
	for _, row := range rows {
		o := Order{
			Object:      "order",
			ID:          row.UUID,
			OrderID:     row.ID,
			Status:      row.Status,
			Payment:     row.Payment,
			Currency:    row.Currency,
			TotalExVAT:  row.TotalExVAT,
			VATTotal:    row.VATTotal,
			TotalIncVAT: row.TotalIncVAT,
			Created:     row.Created,
			Modified:    row.Modified,
		}
		orders = append(orders, &o)
	}
	return orders, nil
}

// GetOrder returns an order by order ID or nil if an error occurred.
func (s *Service) GetOrder(ctx context.Context, orderID string) (*Order, error) {
	contextLogger := log.WithContext(ctx)
	contextLogger.Debugf("service: GetOrder(ctx, orderID=%q", orderID)

	orow, oirows, bill, ship, err := s.model.GetOrderDetailsByUUID(ctx, orderID)
	if err == postgres.ErrOrderNotFound {
		return nil, ErrOrderNotFound
	}
	if err == postgres.ErrOrderItemsNotFound {
		return nil, ErrOrderItemsNotFound
	}
	if err != nil {
		return nil, errors.Wrapf(err,
			"service: s.model.GetOrderDetailsByUUID(ctx, orderUUDID=%q)",
			orderID)
	}

	orderItems := make([]*OrderItem, 0, 8)
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

	return &order, nil
}

// SetStripePaymentIntentID attaches the payment intent id reference to an
// existing order.
func (s *Service) SetStripePaymentIntentID(ctx context.Context, orderID, pi string) error {
	err := s.model.SetStripePaymentIntent(ctx, orderID, pi)
	if err != nil {
		return err
	}
	return nil
}
