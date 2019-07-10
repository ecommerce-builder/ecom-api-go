package postgres

import (
	"context"
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"math"
	"time"

	"github.com/pkg/errors"
)

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
	ID          int
	UUID        string
	Status      string
	Payment     string
	CustomerID  *int
	ShipTb      bool
	Billing     *orderAddress
	Shipping    *orderAddress
	TotalExVAT  int
	TotalVAT    int
	TotalIncVAT int
	Created     time.Time
}

// OrderItemRow holds a single row of data from the order_items table.
type OrderItemRow struct {
	id        int
	UUID      string
	OrderID   int
	SKU       string
	Name      string
	Qty       int
	UnitPrice int
	Discount  *int
	TaxCode   string
	VAT       int
	Created   time.Time
}

// AddOrder adds a new order to the database returning the order row. If
// shipping is nil ship_tb (ship to billing address) is set to true.
// Returns both the OrderRow and list of OrderItemRows as well as the
// order total including VAT to be paid, or nil, nil, 0 if an error occurs.
func (m *PgModel) AddOrder(ctx context.Context, customerUUID *string, cartUUID string, billing, shipping *NewAddress) (*OrderRow, []*OrderItemRow, *CustomerRow, int, error) {
	tx, err := m.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, nil, nil, 0, errors.Wrap(err, "db.BeginTx")
	}

	query := `
		SELECT
		  C.id, C.uuid, C.sku, P.name, qty, unit_price, C.created, C.modified
		FROM carts AS C JOIN products as P
		  ON c.sku = p.sku
		WHERE C.uuid = $1
	`
	rows, err := tx.QueryContext(ctx, query, cartUUID)
	if err != nil {
		return nil, nil, nil, 0, errors.Wrapf(err, "tx.QueryContext(ctx, %q, %q) failed", query, cartUUID)
	}
	defer rows.Close()

	cartItems := make([]*CartProductItem, 0, 20)
	for rows.Next() {
		c := CartProductItem{}
		if err = rows.Scan(&c.id, &c.UUID, &c.SKU, &c.Name, &c.Qty, &c.UnitPrice, &c.Created, &c.Modified); err != nil {
			return nil, nil, nil, 0, errors.Wrapf(err, "scan cart item %v", c)
		}
		cartItems = append(cartItems, &c)
	}
	if err = rows.Err(); err != nil {
		return nil, nil, nil, 0, errors.Wrapf(err, "rows err")
	}

	var customerID *int
	var c CustomerRow
	if customerUUID != nil {
		query = `
			SELECT
			  id, uuid, uid, role, email, firstname, lastname, created, modified
			FROM customers
			WHERE
			  uuid = $1
		`
		if err := tx.QueryRowContext(ctx, query, *customerUUID).Scan(&c.id, &c.UUID, &c.UID, &c.Role, &c.Email, &c.Firstname, &c.Lastname, &c.Created, &c.Modified); err != nil {
			return nil, nil, nil, 0, errors.Wrapf(err, "query row context query=%q", query)
		}
		customerID = &c.id
	}
	fmt.Printf("%#v\n", customerID)

	query = `
		INSERT INTO orders (
			status, payment, customer_id, ship_tb, billing, shipping, total_ex_vat, vat_total, total_inc_vat
		) VALUES (
			'incomplete', 'unpaid', $1, $2, $3, $4, $5, $6, $7
		) RETURNING
			id, uuid, status, payment, customer_id, ship_tb, billing, shipping, total_ex_vat, vat_total, total_inc_vat, created
	`
	var shipTb bool
	if shipping == nil {
		shipTb = true
	} else {
		shipTb = false
	}

	o := OrderRow{}
	totalExVAT, totalVAT := totalSpend(cartItems)
	totalIncVAT := totalExVAT + totalVAT
	row := tx.QueryRowContext(ctx, query, customerID, shipTb, billing, shipping, totalExVAT, totalVAT, totalIncVAT)
	if err := row.Scan(&o.ID, &o.UUID, &o.Status, &o.Payment, &o.CustomerID, &o.ShipTb, &o.Billing, &o.Shipping, &o.TotalExVAT, &o.TotalVAT, &o.TotalIncVAT, &o.Created); err != nil {
		tx.Rollback()
		return nil, nil, nil, 0, errors.Wrapf(err, "tx.QueryRowContext(ctx, %q) failed", query)
	}

	query = `
		INSERT INTO order_items (
		  order_id, sku, name, qty, unit_price, discount, tax_code, vat, created
		) VALUES (
		  $1, $2, $3, $4, $5, $6, $7, $8, NOW()
		) RETURNING
		  id, uuid, order_id, sku, name, qty, unit_price, discount, tax_code, vat, created
	`
	stmt, err := tx.PrepareContext(ctx, query)
	if err != nil {
		tx.Rollback()
		return nil, nil, nil, 0, errors.Wrapf(err, "tx prepare for query=%q", query)
	}
	defer stmt.Close()

	orderItems := make([]*OrderItemRow, 0, len(cartItems))
	for _, t := range cartItems {
		oi := OrderItemRow{}
		row := stmt.QueryRowContext(ctx, o.ID, t.SKU, t.Name, t.Qty, t.UnitPrice, nil, "T20", vat20Normalised(t.Qty*t.UnitPrice))
		if err := row.Scan(&oi.id, &oi.UUID, &oi.OrderID, &oi.SKU, &oi.Name, &oi.Qty, &oi.UnitPrice, &oi.Discount, &oi.TaxCode, &oi.VAT, &oi.Created); err != nil {
			tx.Rollback()
			return nil, nil, nil, 0, errors.Wrap(err, "stmt.QueryRowContext failed")
		}
		orderItems = append(orderItems, &oi)
	}

	if err = tx.Commit(); err != nil {
		return nil, nil, nil, 0, errors.Wrap(err, "tx.Commit() failed")
	}

	return &o, orderItems, &c, totalIncVAT, nil
}
