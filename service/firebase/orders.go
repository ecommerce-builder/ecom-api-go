package firebase

import (
	"context"
	"fmt"
	"time"

	"bitbucket.org/andyfusniakteam/ecom-api-go/model/postgres"
	"github.com/pkg/errors"
)

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

// Order contains details of an previous order.
type Order struct {
	ID       string        `json:"id"`
	Billing  *OrderAddress `json:"billing_address"`
	Shipping *OrderAddress `json:"shipping_address"`
	Created  time.Time     `json:"created"`
}

// PlaceOrder places a new order in the system.
func (s *Service) PlaceOrder(ctx context.Context, contactName, email string, billing *NewAddress, shipping *NewAddress) (*Order, error) {
	customerUUID := "5affdeb2-8e95-4fa7-a248-b4334d0d6d19"
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
	orow, err := s.model.AddOrder(ctx, customerUUID, &pgBilling, &pgShipping)
	if err != nil {
		return nil, errors.Wrapf(err, "s.model.AddOrder(ctx, %q, ...) failed", customerUUID)
	}
	fmt.Printf("%#v\n", orow.Billing)
	order := Order{
		ID: orow.UUID,
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
		Created: orow.Created,
	}
	return &order, nil
}
