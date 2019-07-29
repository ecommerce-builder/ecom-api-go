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

// NewAddress contains the new address request body
type NewAddress struct {
	ContactName string `json:"contact_name"`
	Addr1       string `json:"addr1"`
	Addr2       string `json:"addr2"`
	City        string `json:"city"`
	County      string `json:"county"`
	Postcode    string `json:"postcode"`
	Country     string `json:"country"`
}

// OrderAddress contains details of an address within an Order.
type OrderAddress struct {
	ContactName string  `json:"contact_name"`
	Addr1       string  `json:"addr1"`
	Addr2       *string `json:"addr2,omitempty"`
	City        string  `json:"city"`
	County      *string `json:"county,omitempty"`
	Postcode    string  `json:"postcode"`
	Country     string  `json:"country"`
}

// OrderItem contains details of a line item within an Order.
type OrderItem struct {
	ID        string     `json:"id"`
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

// OrderCustomer contains details of the guest or customer that placed the order.
type OrderCustomer struct {
	ID          string `json:"id,omitempty"`
	ContactName string `json:"contact_name,omitempty"`
	Email       string `json:"email,omitempty"`
}

// Order contains details of an previous order.
type Order struct {
	ID          string         `json:"id"`
	Status      string         `json:"status"`
	Payment     string         `json:"payment"`
	Customer    *OrderCustomer `json:"customer"`
	Billing     *OrderAddress  `json:"billing_address"`
	Shipping    *OrderAddress  `json:"shipping_address"`
	Currency    string         `json:"currency"`
	TotalExVAT  int            `json:"total_ex_vat"`
	VATTotal    int            `json:"vat_total"`
	TotalIncVAT int            `json:"total_inc_vat"`
	Items       []*OrderItem   `json:"items"`
	Created     time.Time      `json:"created"`
	Modified    time.Time      `json:"modified"`
}

// PlaceOrder places a new order in the system.
func (s *Service) PlaceOrder(ctx context.Context, contactName, email *string, customerID *string, cartID string, billing *NewAddress, shipping *NewAddress) (*Order, error) {
	contextLogger := log.WithContext(ctx)
	contextLogger.Debugf("PlaceOrder(ctx, contactName=%v, email=%v, customerID=%v, cartID=%v, ...)", contactName, email, customerID, cartID)

	// Guest orders have a customerUUID of nil whereas Customer orders
	// are set to the UUID of that customer.
	var customerUUID *string
	if customerID != nil {
		customer, err := s.GetCustomer(ctx, *customerID)
		if err != nil {
			if err == ErrCustomerNotFound {
				return nil, err
			}
			return nil, errors.Wrapf(err, "s.GetCustomer(ctx, %q) failed", *customerID)
		}
		customerUUID = customerID
		contextLogger.Debugf("%#v\n", customer)
	} else {
		customerUUID = nil
	}

	if customerUUID == nil {
		contextLogger.Debugf("customerUUID is nil")
	} else {
		contextLogger.Debugf("customerUUID is %s", *customerUUID)
	}

	// Prevent orders with empty carts.
	has, err := s.HasCartItems(ctx, cartID)
	if err != nil {
		return nil, errors.Wrapf(err, "s.HasCartItems(ctx, %q) failed", cartID)
	}

	if !has {
		return nil, ErrCartEmpty
	}

	pgBilling := postgres.NewAddress{
		ContactName: billing.ContactName,
		Addr1:       billing.Addr1,
		Addr2:       billing.Addr2,
		City:        billing.City,
		County:      billing.County,
		Postcode:    billing.Postcode,
		Country:     billing.Country,
	}
	pgShipping := postgres.NewAddress{
		ContactName: shipping.ContactName,
		Addr1:       shipping.Addr1,
		Addr2:       shipping.Addr2,
		City:        shipping.City,
		County:      shipping.County,
		Postcode:    shipping.Postcode,
		Country:     shipping.Country,
	}
	log.WithContext(ctx).Debugf("customerUUID=%v, cartID=%v", customerUUID, cartID)

	orow, oirows, crow, err := s.model.AddOrder(ctx, contactName, email, customerUUID, nil, cartID, &pgBilling, &pgShipping)
	if err != nil {
		return nil, errors.Wrap(err, "s.model.AddOrder(ctx, ...) failed")
	}

	orderItems := make([]*OrderItem, 0, len(oirows))
	for _, oir := range oirows {
		oi := OrderItem{
			ID:        oir.UUID,
			SKU:       oir.SKU,
			Name:      oir.Name,
			Qty:       oir.Qty,
			UnitPrice: oir.UnitPrice,
			Currency:  oir.Currency,
			Discount:  oir.Discount,
			TaxCode:   oir.TaxCode,
			VAT:       oir.VAT,
			Created:   &oir.Created,
		}
		orderItems = append(orderItems, &oi)
	}

	var customer OrderCustomer
	if customerID != nil {
		customer.ID = crow.UUID
	} else {
		customer.ContactName = *contactName
		customer.Email = *email
	}

	order := Order{
		ID:       orow.UUID,
		Status:   orow.Status,
		Payment:  orow.Payment,
		Customer: &customer,
		Billing: &OrderAddress{
			ContactName: orow.Billing.ContactName,
			Addr1:       orow.Billing.Addr1,
			Addr2:       &orow.Billing.Addr2,
			City:        orow.Billing.City,
			County:      &orow.Billing.County,
			Postcode:    orow.Billing.Postcode,
			Country:     orow.Billing.Country,
		},
		Shipping: &OrderAddress{
			ContactName: orow.Shipping.ContactName,
			Addr1:       orow.Shipping.Addr1,
			Addr2:       &orow.Shipping.Addr2,
			City:        orow.Shipping.City,
			County:      &orow.Shipping.County,
			Postcode:    orow.Shipping.Postcode,
			Country:     orow.Shipping.Country,
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
	orow, oirows, err := s.model.GetOrderDetailsByUUID(ctx, orderID)
	if err != nil {
		if err == postgres.ErrOrderNotFound || err == postgres.ErrOrderItemsNotFound {
			return nil, errors.Wrapf(err, "s.model.GetOrderDetailsByUUID(ctx, orderID=%s)", orderID)
		}
		return nil, errors.Wrap(err, "GetOrderDetailsByUUID failed")
	}

	//var customer *Customer
	//crow, err := s.model.GetCustomerByID(ctx, *orow.CustomerID)
	//if err != nil {
	//      if err == postgres.ErrCustomerNotFound {
	//              customer = nil
	//      } else {
	//              return nil, errors.Wrapf(err, "s.model.GetCustomerByID(ctx, CustomerID=%s", *orow.CustomerID)
	//      }
	//}
	//if crow != nil {
	//      customer = &Customer{
	//              ID:        crow.UUID,
	//              UID:       crow.UID,
	//              Role:      crow.Role,
	//              Email:     crow.Email,
	//              Firstname: crow.Firstname,
	//              Lastname:  crow.Lastname,
	//              Created:   crow.Created,
	//              Modified:  crow.Modified,
	//      }
	//}

	orderItems := make([]*OrderItem, 0, 8)
	for _, oir := range oirows {
		oi := OrderItem{
			ID:        oir.UUID,
			SKU:       oir.SKU,
			Name:      oir.Name,
			Qty:       oir.Qty,
			UnitPrice: oir.UnitPrice,
			Currency:  oir.Currency,
			Discount:  oir.Discount,
			TaxCode:   oir.TaxCode,
			VAT:       oir.VAT,
			Created:   nil,
		}
		orderItems = append(orderItems, &oi)
	}
	order := Order{
		ID:       orow.UUID,
		Status:   orow.Status,
		Payment:  orow.Payment,
		Currency: "GBP",
		//Customer: customer,
		Billing: &OrderAddress{
			ContactName: orow.Billing.ContactName,
			Addr1:       orow.Billing.Addr1,
			Addr2:       &orow.Billing.Addr2,
			City:        orow.Billing.City,
			County:      &orow.Billing.County,
			Postcode:    orow.Billing.Postcode,
			Country:     orow.Billing.Country,
		},
		Shipping: &OrderAddress{
			ContactName: orow.Shipping.ContactName,
			Addr1:       orow.Shipping.Addr1,
			Addr2:       &orow.Shipping.Addr2,
			City:        orow.Shipping.City,
			County:      &orow.Shipping.County,
			Postcode:    orow.Shipping.Postcode,
			Country:     orow.Shipping.Country,
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
