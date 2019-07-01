package postgres

import (
	"context"
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"

	"github.com/pkg/errors"
)

// OrderAddress holds the JSONB field of the OrderRow shipping and billing columns.
type OrderAddress struct {
	ContactName string
	Addr1       string
	Addr2       string
	City        string
	County      string
	Postcode    string
	Country     string
}

// Scan unmarshals JSON data into a ProductContent struct
func (oa *OrderAddress) Scan(value interface{}) error {
	sv, err := driver.String.ConvertValue(value)
	if err != nil {
		return errors.Wrap(err, "convert value failed")
	}
	if v, ok := sv.([]byte); ok {
		var orderAddr OrderAddress
		err := json.Unmarshal(v, &orderAddr)
		if err != nil {
			return errors.Wrap(err, "json unmarshal of OrderAddress failed")
		}
		*oa = orderAddr
		return nil
	}
	return fmt.Errorf("scan value failed")
}

// OrderRow holds a single row of data from the orders table.
type OrderRow struct {
	ID         int
	UUID       string
	CustomerID int
	ShipTb     bool
	Billing    *OrderAddress
	Shipping   *OrderAddress
	Total      int
	Created    time.Time
}

// AddOrder adds a new order to the database returning the order row. If
// shipping is nil ship_tb (ship to billing address) is set to true.
func (m *PgModel) AddOrder(ctx context.Context, customerUUID string, billing, shipping *NewAddress) (*OrderRow, error) {
	o := OrderRow{}
	query := `
		INSERT INTO orders (
			customer_id, ship_tb, billing, shipping, total
		) VALUES (
			$1, $2, $3, $4, $5
		) RETURNING
			id, uuid, customer_id, ship_tb, billing, shipping, total, created
	`
	customerID := 5
	total := 0
	var shipTb bool
	if shipping == nil {
		shipTb = true
	} else {
		shipTb = false
	}
	err := m.db.QueryRowContext(ctx, query, customerID, shipTb, billing, shipping, total).
		Scan(&o.ID, &o.UUID, &o.CustomerID, &o.ShipTb, &o.Billing, &o.Shipping, &o.Total, &o.Created)
	if err != nil {
		return nil, err
	}
	return &o, nil
}
