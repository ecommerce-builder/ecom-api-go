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
	ID          string `json:"id,omitempty"`
	ContactName string `json:"contact_name,omitempty"`
	Email       string `json:"email,omitempty"`
}

// Order contains details of an previous order.
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
		User:    nil,
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
			ID: urow.UUID,
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

// GetOrder returns an order by order ID or nil if an error occurred.
func (s *Service) GetOrder(ctx context.Context, orderID string) (*Order, error) {
	contextLogger := log.WithContext(ctx)
	contextLogger.Debugf("service: GetOrder(ctx, orderID=%q", orderID)

	orow, oirows, bill, ship, err := s.model.GetOrderDetailsByUUID(ctx, orderID)
	if err == postgres.ErrOrderNotFound || err == postgres.ErrOrderItemsNotFound {
		return nil, errors.Wrapf(err, "s.model.GetOrderDetailsByUUID(ctx, orderID=%s)", orderID)
	}
	if err != nil {
		return nil, errors.Wrap(err, "GetOrderDetailsByUUID failed")
	}

	orderItems := make([]*OrderItem, 0, 8)
	for _, v := range oirows {
		oi := OrderItem{
			ID:        v.UUID,
			Path:      v.Path,
			SKU:       v.SKU,
			Name:      v.Name,
			Qty:       v.Qty,
			UnitPrice: v.UnitPrice,
			Currency:  v.Currency,
			Discount:  v.Discount,
			TaxCode:   v.TaxCode,
			VAT:       v.VAT,
			Created:   nil,
		}
		orderItems = append(orderItems, &oi)
	}
	order := Order{
		ID:       orow.UUID,
		OrderID:  orow.ID,
		Status:   orow.Status,
		Payment:  orow.Payment,
		Currency: orow.Currency,
		//User: user,
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
		Items:   orderItems,
		Created: orow.Created,
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
