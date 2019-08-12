package postgres

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"math"
	"time"

	"github.com/pkg/errors"
)

// ErrOrderNotFound is returned by when the order with
// the given ID could not be found in the database.
var ErrOrderNotFound = errors.New("model: order not found")

// ErrOrderItemsNotFound is returned by when the order item
// associated to an order could not be found in the database.
var ErrOrderItemsNotFound = errors.New("model: order item not found")

// orderAddress holds the JSONB field of the OrderRow shipping and billing columns.
type orderAddress struct {
	ContactName string
	Addr1       string
	Addr2       string
	City        string
	County      string
	Postcode    string
	Country     string
}

// Scan unmarshals JSON data into a ProductContent struct
func (oa *orderAddress) Scan(value interface{}) error {
	sv, err := driver.String.ConvertValue(value)
	if err != nil {
		return errors.Wrap(err, "convert value failed")
	}
	if v, ok := sv.([]byte); ok {
		var orderAddr orderAddress
		err := json.Unmarshal(v, &orderAddr)
		if err != nil {
			return errors.Wrap(err, "json unmarshal of orderAddress failed")
		}
		*oa = orderAddr
		return nil
	}
	return fmt.Errorf("scan value failed")
}

func vat20Normalised(price int) int {
	return int(math.Round(float64(price) * 0.2))
}

func discountMultiplier(disc int) float64 {
	if disc < 0 || disc > 10000 {
		return 0
	}
	return float64(10000-disc) / 10000.0
}

func totalSpend(cartItems []*CartProductItem) (int, int) {
	totalExVAT := 0
	totalVAT := 0
	for _, i := range cartItems {
		totalExVAT = totalExVAT + (i.Qty * i.UnitPrice)
		totalVAT = totalVAT + vat20Normalised(i.Qty*i.UnitPrice)
	}
	return totalExVAT, totalVAT
}

// OrderRow holds a single row of data from the orders table.
type OrderRow struct {
	id            int
	UUID          string
	Status        string
	Payment       string
	CustomerID    *int
	CustomerName  *string
	CustomerEmail *string
	StripePI      *string
	ShipTb        bool
	Billing       *orderAddress
	Shipping      *orderAddress
	Currency      string
	TotalExVAT    int
	VATTotal      int
	TotalIncVAT   int
	Created       time.Time
	Modified      time.Time
}

// OrderItemRow holds a single row of data from the order_item table.
type OrderItemRow struct {
	id        int
	UUID      string
	orderID   int
	SKU       string
	Name      string
	Qty       int
	UnitPrice int
	Currency  string
	Discount  *int
	TaxCode   string
	VAT       int
	Created   time.Time
}

// AddOrder adds a new order to the database returning the order row. If
// shipping is nil ship_tb (ship to billing address) is set to true.
// Returns both the OrderRow and list of OrderItemRows as well as the
// order total including VAT to be paid, or nil, nil, 0 if an error occurs.
func (m *PgModel) AddOrder(ctx context.Context, customerName, customerEmail, customerUUID *string, stripePI *string, cartUUID string, billing, shipping *NewAddress) (*OrderRow, []*OrderItemRow, *CustomerRow, error) {
	tx, err := m.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, nil, nil, errors.Wrap(err, "db.BeginTx")
	}

	query := `
		SELECT
		  C.id, C.uuid, C.sku, P.name, qty, unit_price, C.created, C.modified
		FROM cart AS C JOIN product as P
		  ON c.sku = p.sku
		WHERE C.uuid = $1
	`
	rows, err := tx.QueryContext(ctx, query, cartUUID)
	if err != nil {
		return nil, nil, nil, errors.Wrapf(err, "tx.QueryContext(ctx, %q, %q) failed", query, cartUUID)
	}
	defer rows.Close()

	cartItems := make([]*CartProductItem, 0, 20)
	for rows.Next() {
		c := CartProductItem{}
		if err = rows.Scan(&c.id, &c.UUID, &c.SKU, &c.Name, &c.Qty, &c.UnitPrice, &c.Created, &c.Modified); err != nil {
			return nil, nil, nil, errors.Wrapf(err, "scan cart item %v", c)
		}
		cartItems = append(cartItems, &c)
	}
	if err = rows.Err(); err != nil {
		return nil, nil, nil, errors.Wrapf(err, "rows err")
	}

	var customerID *int
	var c CustomerRow
	if customerUUID != nil {
		query = `
			SELECT
			  id, uuid, uid, role, email, firstname, lastname, created, modified
			FROM customer
			WHERE
			  uuid = $1
		`
		if err := tx.QueryRowContext(ctx, query, *customerUUID).Scan(&c.id, &c.UUID, &c.UID, &c.Role, &c.Email, &c.Firstname, &c.Lastname, &c.Created, &c.Modified); err != nil {
			return nil, nil, nil, errors.Wrapf(err, "query row context query=%q", query)
		}
		customerID = &c.id
	}

	query = `
		INSERT INTO order (
			status, payment, customer_id, customer_name, customer_email, stripe_pi, ship_tb,
			billing, shipping, currency, total_ex_vat, vat_total, total_inc_vat,
			created, modified
		) VALUES (
			'incomplete', 'unpaid', $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, NOW(), NOW()
		) RETURNING
			id, uuid, status, payment, customer_id, customer_name, customer_email, stripe_pi,
			ship_tb, billing, shipping, currency, total_ex_vat, vat_total,
			total_inc_vat, created, modified
	`
	var shipTb bool
	if shipping == nil {
		shipTb = true
	} else {
		shipTb = false
	}

	o := OrderRow{}
	currency := "GBP" // hardcoded for now but may come from elsewhere later.
	totalExVAT, totalVAT := totalSpend(cartItems)
	totalIncVAT := totalExVAT + totalVAT

	row := tx.QueryRowContext(ctx, query, customerID, customerName, customerEmail, stripePI, shipTb, billing, shipping, currency, totalExVAT, totalVAT, totalIncVAT)
	if err := row.Scan(&o.id, &o.UUID, &o.Status, &o.Payment, &o.CustomerID, &o.CustomerName, &o.CustomerEmail, &o.StripePI,
		&o.ShipTb, &o.Billing, &o.Shipping, &o.Currency, &o.TotalExVAT, &o.VATTotal,
		&o.TotalIncVAT, &o.Created, &o.Modified); err != nil {
		tx.Rollback()
		return nil, nil, nil, errors.Wrapf(err, "tx.QueryRowContext(ctx, %q) failed", query)
	}

	query = `
		INSERT INTO order_item (
		  order_id, sku, name, qty, unit_price, discount, tax_code, vat, created
		) VALUES (
		  $1, $2, $3, $4, $5, $6, $7, $8, NOW()
		) RETURNING
		  id, uuid, order_id, sku, name, qty, unit_price, currency, discount, tax_code, vat, created
	`
	stmt, err := tx.PrepareContext(ctx, query)
	if err != nil {
		tx.Rollback()
		return nil, nil, nil, errors.Wrapf(err, "tx prepare for query=%q", query)
	}
	defer stmt.Close()

	orderItems := make([]*OrderItemRow, 0, len(cartItems))
	for _, t := range cartItems {
		oi := OrderItemRow{}

		row := stmt.QueryRowContext(ctx, o.id, t.SKU, t.Name, t.Qty, t.UnitPrice, nil, "T20", vat20Normalised(t.Qty*t.UnitPrice))
		if err := row.Scan(&oi.id, &oi.UUID, &oi.orderID, &oi.SKU, &oi.Name, &oi.Qty, &oi.UnitPrice, &oi.Currency, &oi.Discount, &oi.TaxCode, &oi.VAT, &oi.Created); err != nil {
			tx.Rollback()
			return nil, nil, nil, errors.Wrap(err, "stmt.QueryRowContext failed")
		}
		orderItems = append(orderItems, &oi)
	}

	if err = tx.Commit(); err != nil {
		return nil, nil, nil, errors.Wrap(err, "tx.Commit() failed")
	}

	return &o, orderItems, &c, nil
}

// GetOrderDetailsByUUID retrieves the order row and order item rows
// for a given order. If the order cannot be found GetOrderDetailsByUUID
// returns an error == ErrOrderNotFound.
func (m *PgModel) GetOrderDetailsByUUID(ctx context.Context, orderUUID string) (*OrderRow, []*OrderItemRow, error) {
	tx, err := m.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, nil, errors.Wrap(err, "db.BeginTx")
	}

	query := `
		SELECT
		  id,  uuid, status, payment, customer_id, customer_name, customer_email, stripe_pi,
		  ship_tb, billing, shipping, currency, total_ex_vat, vat_total,
		  total_inc_vat, created, modified
		FROM order
		WHERE uuid = $1
	`
	o := OrderRow{}
	if err := tx.QueryRowContext(ctx, query, orderUUID).Scan(&o.id, &o.UUID, &o.Status, &o.Payment, &o.CustomerID, &o.CustomerName, &o.CustomerEmail, &o.StripePI,
		&o.ShipTb, &o.Billing, &o.Shipping, &o.Currency, &o.TotalExVAT, &o.VATTotal,
		&o.TotalIncVAT, &o.Created, &o.Modified); err != nil {
		if err == sql.ErrNoRows {
			tx.Rollback()
			return nil, nil, ErrOrderNotFound
		}
		tx.Rollback()
		return nil, nil, errors.Wrapf(err, "query row context query=%q", query)
	}
	query = `
		SELECT
		  id, uuid, order_id, sku, name, qty, unit_price, currency, discount, tax_code, vat, created
		FROM order_item
		WHERE order_id = $1
	`
	rows, err := tx.QueryContext(ctx, query, o.id)
	if err != nil {
		if err == sql.ErrNoRows {
			tx.Rollback()
			return nil, nil, ErrOrderItemsNotFound
		}
		tx.Rollback()
		return nil, nil, errors.Wrapf(err, "tx.QueryContext(ctx, %q, %q) failed", query, o.id)
	}
	defer rows.Close()
	orderItems := make([]*OrderItemRow, 0, 16)
	for rows.Next() {
		i := OrderItemRow{}
		if err = rows.Scan(&i.id, &i.UUID, &i.orderID, &i.SKU, &i.Name, &i.Qty, &i.UnitPrice, &i.Currency, &i.Discount, &i.TaxCode, &i.VAT, &i.Created); err != nil {
			return nil, nil, errors.Wrapf(err, "scan order item %v", i)
		}
		orderItems = append(orderItems, &i)
	}
	if err = rows.Err(); err != nil {
		tx.Rollback()
		return nil, nil, errors.Wrapf(err, "rows err")
	}
	if err = tx.Commit(); err != nil {
		tx.Rollback()
		return nil, nil, errors.Wrap(err, "tx.Commit() failed")
	}
	return &o, orderItems, nil
}

// SetStripePaymentIntent sets payment intent id reference on an
// existing order and updates the modified timestamp.
func (m *PgModel) SetStripePaymentIntent(ctx context.Context, orderID, pi string) error {
	query := `
		UPDATE order
		SET stripe_pi = $1, modified = NOW()
		WHERE uuid = $2
	`
	_, err := m.db.ExecContext(ctx, query, pi, orderID)
	if err != nil {
		return errors.Wrapf(err, "m.db.ExecContext(ctx, query=%q, pi=%q, orderID=%q) failed", query, pi, orderID)
	}
	return nil
}

// RecordPayment marks the order with the given order ID and Stripe Intent
// referenceas complete and paid.
func (m *PgModel) RecordPayment(ctx context.Context, orderID, pi string, body []byte) error {
	tx, err := m.db.BeginTx(ctx, nil)
	if err != nil {
		return errors.Wrap(err, "db.BeginTx")
	}

	query := `
		UPDATE order
		SET status = 'completed', payment = 'paid', modified = NOW()
		WHERE uuid = $1 AND stripe_pi = $2
		RETURNING id
	`
	var id int
	err = tx.QueryRowContext(ctx, query, orderID, pi).Scan(&id)
	if err != nil {
		return errors.Wrapf(err, "tx.ExecContext(ctx, query=%q, pi=%q, orderID=%q) failed", query, pi, orderID)
	}

	query = `
		INSERT INTO payment (
		  order_id, typ, result, created
		) VALUES (
		  $1, 'stripe', $2, NOW()
		)
	`
	_, err = tx.ExecContext(ctx, query, id, body)
	if err != nil {
		return errors.Wrapf(err, "tx.ExecContext(ctx, query=%q, id=%d, body=%q) failed", query, id, body)
	}

	if err = tx.Commit(); err != nil {
		tx.Rollback()
		return errors.Wrap(err, "tx.Commit() failed")
	}

	return nil
}
