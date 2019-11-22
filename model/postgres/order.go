package postgres

import (
	"context"
	"database/sql"
	"math"
	"time"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

// ErrOrderNotFound is returned by when the order with
// the given ID could not be found in the database.
var ErrOrderNotFound = errors.New("postgres: order not found")

// ErrOrderItemsNotFound is returned by when the order item
// associated to an order could not be found in the database.
var ErrOrderItemsNotFound = errors.New("postgres: order item not found")

// ErrCartEmpty error
var ErrCartEmpty = errors.New("postgres: cart not found")

// Scan unmarshals JSON data into a ProductContent struct
// func (oa *orderAddress) Scan(value interface{}) error {
// 	sv, err := driver.String.ConvertValue(value)
// 	if err != nil {
// 		return errors.Wrap(err, "convert value failed")
// 	}
// 	if v, ok := sv.([]byte); ok {
// 		var orderAddr orderAddress
// 		err := json.Unmarshal(v, &orderAddr)
// 		if err != nil {
// 			return errors.Wrap(err, "json unmarshal of orderAddress failed")
// 		}
// 		*oa = orderAddr
// 		return nil
// 	}
// 	return fmt.Errorf("scan value failed")
// }

func vat20Normalised(price int) int {
	return int(math.Round(float64(price) * 0.2))
}

func discountMultiplier(disc int) float64 {
	if disc < 0 || disc > 10000 {
		return 0
	}
	return float64(10000-disc) / 10000.0
}

func totalSpend(cartProducts []*CartProductJoinRow) (int, int) {
	totalExVAT := 0
	totalVAT := 0
	for _, i := range cartProducts {
		totalExVAT = totalExVAT + (i.Qty * i.UnitPrice)
		totalVAT = totalVAT + vat20Normalised(i.Qty*i.UnitPrice)
	}
	return totalExVAT, totalVAT
}

// NewOrderAddress object
type NewOrderAddress struct {
	ContactName string
	Addr1       string
	Addr2       *string
	City        string
	County      *string
	Postcode    string
	CountryCode string
}

// OrderRow holds a single row of data from the orders table.
type OrderRow struct {
	id          int
	UUID        string
	usrID       *int
	Status      string
	Payment     string
	ContactName *string
	Email       *string
	StripePI    *string
	billingID   int
	shippingID  int
	Currency    string
	TotalExVAT  int
	VATTotal    int
	TotalIncVAT int
	Created     time.Time
	Modified    time.Time
}

// OrderItemRow holds a single row of data from the order_item table.
type OrderItemRow struct {
	id        int
	UUID      string
	orderID   int
	Path      string
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

// OrderAddressRow holds a single row of data from the order_address table.
type OrderAddressRow struct {
	id          int
	UUID        string
	Typ         string
	ContactName string
	Addr1       string
	Addr2       *string
	City        string
	County      *string
	Postcode    string
	CountryCode string
	Created     time.Time
	Modified    time.Time
}

// AddGuestOrder adds a new guest order to the database returning the order row,
// slice of order item rows, a billing and shipping address row.
func (m *PgModel) AddGuestOrder(ctx context.Context, cartUUID, contactName, email string,
	billing, shipping *NewOrderAddress) (*OrderRow, []*OrderItemRow, *OrderAddressRow, *OrderAddressRow, error) {
	contextLogger := log.WithContext(ctx)
	contextLogger.Debugf("postgres: AddGuestOrder(ctx, cartUUID=%q, contactName=%s, email=%s, ...)",
		cartUUID, contactName, email)

	// start transaction
	tx, err := m.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, nil, nil, nil,
			errors.Wrap(err, "postgres: db.BeginTx")
	}

	// 1. Check the cart exists
	q1 := "SELECT id FROM cart WHERE uuid = $1"
	var cartID int
	err = tx.QueryRowContext(ctx, q1, cartUUID).Scan(&cartID)
	if err == sql.ErrNoRows {
		tx.Rollback()
		return nil, nil, nil, nil, ErrCartNotFound
	}
	if err != nil {
		tx.Rollback()
		return nil, nil, nil, nil, errors.Wrapf(err,
			"postgres: query row context failed for q1=%q", q1)
	}

	contextLogger.Debugf("postgres: q1 returned a cartID=%d", cartID)

	// 2. Get a list all all products in the cart with their price.
	q2 := `
		SELECT
		  c.id, c.uuid,
		  p.id, p.uuid,
		  p.path, p.sku, p.name, qty,
		  r.unit_price,
		  c.created, c.modified
		FROM cart_product AS c
		JOIN product AS p
		  ON c.product_id = p.id
		JOIN price AS r
		  ON c.product_id = r.product_id
		WHERE r.break = 1 AND c.cart_id = $1
	`
	rows, err := tx.QueryContext(ctx, q2, cartID)
	if err != nil {
		return nil, nil, nil, nil, errors.Wrapf(err,
			"postgres: tx.QueryContext(ctx, q2=%q, cartID=%d) failed",
			q2, cartID)
	}
	defer rows.Close()

	cartProducts := make([]*CartProductJoinRow, 0, 20)
	for rows.Next() {
		c := CartProductJoinRow{}
		err = rows.Scan(&c.id, &c.UUID,
			&c.productID, &c.ProductUUID,
			&c.Path, &c.SKU, &c.Name, &c.Qty,
			&c.UnitPrice,
			&c.Created, &c.Modified)
		if err != nil {
			return nil, nil, nil, nil,
				errors.Wrapf(err, "postgres: scan cart item %v", c)
		}
		c.CartUUID = cartUUID
		cartProducts = append(cartProducts, &c)
	}
	if err := rows.Err(); err != nil {
		return nil, nil, nil, nil,
			errors.Wrapf(err, "postgres: rows err")
	}
	contextLogger.Infof("postgres: q2 returned %d products in this cart", len(cartProducts))

	// 3. Insert the billing and shipping addresses.
	q3 := `
		INSERT INTO order_address (
		  typ, contact_name,
		  addr1, addr2, city, county,
		  postcode, country_code,
		  created, modified
		) VALUES (
		  $1, $2,
		  $3, $4, $5, $6,
		  $7, $8,
		  NOW(), NOW()
		) RETURNING
		  id, uuid, typ, contact_name,
		  addr1, addr2, city, county,
		  postcode, country_code,
		  created, modified
	`
	stmt3, err := tx.PrepareContext(ctx, q3)
	if err != nil {
		tx.Rollback()
		return nil, nil, nil, nil, errors.Wrapf(err,
			"postgres: tx prepare for q3=%q", q3)
	}
	defer stmt3.Close()

	var bv OrderAddressRow
	row := stmt3.QueryRowContext(ctx, "billing", billing.ContactName,
		billing.Addr1, billing.Addr2, billing.City, billing.County,
		billing.Postcode, billing.CountryCode)
	err = row.Scan(&bv.id, &bv.UUID, &bv.Typ, &bv.ContactName,
		&bv.Addr1, &bv.Addr2, &bv.City, &bv.County,
		&bv.Postcode, &bv.CountryCode,
		&bv.Created, &bv.Modified)
	if err != nil {
		tx.Rollback()
		return nil, nil, nil, nil, errors.Wrap(err, "postgres: scan failed")
	}

	var sv OrderAddressRow
	row = stmt3.QueryRowContext(ctx, "shipping", shipping.ContactName,
		shipping.Addr1, shipping.Addr2, shipping.City, shipping.County,
		shipping.Postcode, shipping.CountryCode)
	err = row.Scan(&sv.id, &sv.UUID, &sv.Typ, &sv.ContactName,
		&sv.Addr1, &sv.Addr2, &sv.City, &sv.County,
		&sv.Postcode, &sv.CountryCode,
		&sv.Created, &sv.Modified)
	if err != nil {
		tx.Rollback()
		return nil, nil, nil, nil, errors.Wrap(err, "postgres: scan failed")
	}

	// 4. Insert the order row
	q4 := `
		INSERT INTO "order" (
		  status, payment, contact_name, email,
		  billing_id, shipping_id, currency, total_ex_vat,
		  vat_total, total_inc_vat,
		  created, modified
		) VALUES (
		  'incomplete', 'unpaid', $1, $2,
		  $3, $4, $5, $6, $7, $8,
		  NOW(), NOW()
		) RETURNING
		  id, uuid, usr_id, status, payment, contact_name, email, stripe_pi,
		  billing_id, shipping_id, currency, total_ex_vat, vat_total,
		  total_inc_vat, created, modified
	`

	o := OrderRow{}
	currency := "GBP" // hardcoded for now but may come from elsewhere later.
	totalExVAT, totalVAT := totalSpend(cartProducts)
	totalIncVAT := totalExVAT + totalVAT

	row = tx.QueryRowContext(ctx, q4, contactName, email,
		bv.id, sv.id, currency, totalExVAT,
		totalVAT, totalIncVAT)
	err = row.Scan(&o.id, &o.UUID, &o.usrID, &o.Status, &o.Payment,
		&o.ContactName, &o.Email, &o.StripePI, &o.billingID,
		&o.shippingID, &o.Currency, &o.TotalExVAT, &o.VATTotal,
		&o.TotalIncVAT, &o.Created, &o.Modified)
	if err != nil {
		tx.Rollback()
		return nil, nil, nil, nil, errors.Wrapf(err,
			"postgres: tx.QueryRowContext(ctx, q4=%q) failed", q4)
	}

	// 5. Prepared statement for individual order items.
	q5 := `
		INSERT INTO order_item (
		  order_id, path, sku, name,
		  qty, unit_price, discount, tax_code, vat, created
		) VALUES (
		  $1, $2, $3, $4, $5, $6, $7, $8, $9, NOW()
		) RETURNING
		  id, uuid, order_id, sku, name, qty, unit_price, currency, discount, tax_code, vat, created
	`
	stmt, err := tx.PrepareContext(ctx, q5)
	if err != nil {
		tx.Rollback()
		return nil, nil, nil, nil,
			errors.Wrapf(err, "postgres: tx prepare for q5=%q", q5)
	}
	defer stmt.Close()

	orderItems := make([]*OrderItemRow, 0, len(cartProducts))
	for _, t := range cartProducts {
		oi := OrderItemRow{}
		row := stmt.QueryRowContext(ctx, o.id, t.Path, t.SKU, t.Name,
			t.Qty, t.UnitPrice, nil, "T20",
			vat20Normalised(t.Qty*t.UnitPrice))
		err := row.Scan(&oi.id, &oi.UUID, &oi.orderID, &oi.SKU,
			&oi.Name, &oi.Qty, &oi.UnitPrice, &oi.Currency,
			&oi.Discount, &oi.TaxCode, &oi.VAT, &oi.Created)
		if err != nil {
			tx.Rollback()
			return nil, nil, nil, nil,
				errors.Wrap(err, "postgres: stmt.QueryRowContext failed")
		}
		orderItems = append(orderItems, &oi)
	}
	if err := tx.Commit(); err != nil {
		return nil, nil, nil, nil,
			errors.Wrap(err, "postgres: tx.Commit() failed")
	}

	return &o, orderItems, &bv, &sv, nil
}

// AddOrder adds a new order to the database returning the order row. If
// shipping is nil ship_tb (ship to billing address) is set to true.
// Returns both the OrderRow and list of OrderItemRows as well as the
// order total including VAT to be paid, or nil, nil, 0 if an error occurs.
func (m *PgModel) AddOrder(ctx context.Context, cartUUID, userUUID, billingUUID, shippingUUID string) (*OrderRow, []*OrderItemRow, *UsrRow, *OrderAddressRow, *OrderAddressRow, error) {
	contextLogger := log.WithContext(ctx)
	contextLogger.Debugf("postgres: AddOrder(ctx, cartUUID=%q, userUUID=%q, billingUUID=%q, shippingUUID=%q, ...)",
		cartUUID, userUUID, billingUUID, shippingUUID)

	// start transaction
	tx, err := m.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, nil, nil, nil, nil,
			errors.Wrap(err, "postgres: db.BeginTx")
	}

	// 1. Check the cart exists
	q1 := "SELECT id FROM cart WHERE uuid = $1"
	var cartID int
	err = tx.QueryRowContext(ctx, q1, cartUUID).Scan(&cartID)
	if err == sql.ErrNoRows {
		tx.Rollback()
		return nil, nil, nil, nil, nil, ErrCartNotFound
	}
	if err != nil {
		tx.Rollback()
		return nil, nil, nil, nil, nil, errors.Wrapf(err,
			"postgres: query row context failed for q1=%q", q1)
	}

	// 2. Get a list all all products in the cart with their price.
	q2 := `
		SELECT
		  c.id, c.uuid, p.name, qty,
		  r.unit_price,
		  c.created, c.modified
		FROM cart_product AS c
		JOIN product AS p
		  ON c.product_id = p.id
		JOIN price AS r
		  ON c.product_id = r.product_id
		WHERE r.break = 1 AND c.cart_id = $1
	`
	rows, err := tx.QueryContext(ctx, q2, cartID)
	if err != nil {
		return nil, nil, nil, nil, nil, errors.Wrapf(err,
			"postgres: tx.QueryContext(ctx, q2=%q, cartID=%d) failed",
			q2, cartID)
	}
	defer rows.Close()

	cartProducts := make([]*CartProductJoinRow, 0, 20)
	for rows.Next() {
		c := CartProductJoinRow{}
		if err = rows.Scan(&c.id, &c.UUID, &c.Name, &c.Qty,
			&c.UnitPrice, &c.Created, &c.Modified); err != nil {
			return nil, nil, nil, nil, nil,
				errors.Wrapf(err, "postgres: scan cart item %v", c)
		}
		cartProducts = append(cartProducts, &c)
	}
	if err := rows.Err(); err != nil {
		return nil, nil, nil, nil, nil,
			errors.Wrapf(err, "postgres: rows err")
	}

	contextLogger.Infof("postgres: %d products in this cart", len(cartProducts))

	// 3. Get the user
	var c UsrRow
	q3 := `
		SELECT
		  id, uuid, uid, role, email, firstname, lastname, created, modified
		FROM usr
		WHERE
		  uuid = $1
	`
	err = tx.QueryRowContext(ctx, q3, userUUID).Scan(&c.id, &c.UUID,
		&c.UID, &c.Role, &c.Email, &c.Firstname,
		&c.Lastname, &c.Created, &c.Modified)
	if err == sql.ErrNoRows {
		tx.Rollback()
		return nil, nil, nil, nil, nil, ErrUserNotFound
	}
	if err != nil {
		return nil, nil, nil, nil, nil,
			errors.Wrapf(err, "postgres: query row context q3=%q", q3)
	}
	// userID = &c.id

	// 4. Get the billing and shipping addresses
	// and make sure this user owns them.
	q4 := `
		SELECT
		  id, uuid, usr_id, typ,
		  contact_name, addr1, addr2,
		  city, county, postcode,
		  country_code, created, modified
		FROM address
		WHERE uuid = $1 AND usr_id = $2
	`
	stmt4, err := tx.PrepareContext(ctx, q4)
	if err != nil {
		tx.Rollback()
		return nil, nil, nil, nil, nil, errors.Wrapf(err,
			"postgres: tx prepare for q4=%q", q4)
	}
	defer stmt4.Close()

	var abv AddressJoinRow
	row := stmt4.QueryRowContext(ctx, billingUUID, c.id)
	err = row.Scan(&abv.id, &abv.UUID, &abv.usrID, &abv.Typ,
		&abv.ContactName, &abv.Addr1, &abv.Addr2,
		&abv.City, &abv.County, &abv.Postcode,
		&abv.CountryCode, &abv.Created, &abv.Modified)
	if err == sql.ErrNoRows {
		tx.Rollback()
		return nil, nil, nil, nil, nil, ErrAddressNotFound
	}
	if err != nil {
		tx.Rollback()
		return nil, nil, nil, nil, nil,
			errors.Wrap(err, "postgres: scan failed")
	}

	var asv AddressJoinRow
	row = stmt4.QueryRowContext(ctx, shippingUUID, c.id)
	err = row.Scan(&asv.id, &asv.UUID, &asv.usrID, &asv.Typ,
		&asv.ContactName, &asv.Addr1, &asv.Addr2,
		&asv.City, &asv.County, &asv.Postcode,
		&asv.CountryCode, &asv.Created, &asv.Modified)
	if err == sql.ErrNoRows {
		tx.Rollback()
		return nil, nil, nil, nil, nil, ErrAddressNotFound
	}
	if err != nil {
		tx.Rollback()
		return nil, nil, nil, nil, nil,
			errors.Wrap(err, "postgres: scan failed")
	}

	// 5. Insert the billing and shipping addresses.
	q5 := `
		INSERT INTO order_address (
		  typ, contact_name, addr1, addr2, city, county,
		  postcode, country_code,
		  created, modified
		) VALUES (
		  $1, $2, $3, $4, $5, $6,
		  $7, $8,
		  NOW(), NOW()
		)
		RETURNING
		  id, uuid, typ, contact_name, addr1, addr2, city, county,
		  postcode, country_code, created, modified
	`
	stmt5, err := tx.PrepareContext(ctx, q5)
	if err != nil {
		tx.Rollback()
		return nil, nil, nil, nil, nil, errors.Wrapf(err,
			"postgres: tx prepare for q5=%q", q5)
	}
	defer stmt5.Close()

	var bv OrderAddressRow
	row = stmt5.QueryRowContext(ctx, "billing",
		abv.ContactName, abv.Addr1, abv.Addr2,
		abv.City, abv.County, abv.Postcode, abv.CountryCode)
	err = row.Scan(&bv.id, &bv.UUID, &bv.Typ, &bv.ContactName, &bv.Addr1,
		&bv.Addr2, &bv.City, &bv.County, &bv.Postcode, &bv.CountryCode,
		&bv.Created, &bv.Modified)
	if err != nil {
		tx.Rollback()
		return nil, nil, nil, nil, nil,
			errors.Wrap(err, "postgres: scan failed")
	}

	var sv OrderAddressRow
	row = stmt5.QueryRowContext(ctx, "shipping",
		asv.ContactName, asv.Addr1, asv.Addr2,
		asv.City, asv.County, asv.Postcode, asv.CountryCode)
	err = row.Scan(&sv.id, &sv.UUID, &sv.Typ, &sv.ContactName, &sv.Addr1,
		&sv.Addr2, &sv.City, &sv.County, &sv.Postcode, &sv.CountryCode,
		&sv.Created, &sv.Modified)
	if err != nil {
		tx.Rollback()
		return nil, nil, nil, nil, nil,
			errors.Wrap(err, "postgres: scan failed")
	}

	// 6. Insert the order
	q6 := `
		INSERT INTO "order" (
		  status, payment, usr_id,
		  billing_id, shipping_id, currency,
		  total_ex_vat, vat_total, total_inc_vat,
		  created, modified
		) VALUES (
		  'incomplete', 'unpaid', $1,
		  $2, $3, $4,
		  $5, $6, $7,
		  NOW(), NOW()
		) RETURNING
		  id, uuid, usr_id, status, payment, contact_name, email, stripe_pi,
		  billing_id, shipping_id, currency, total_ex_vat, vat_total,
		  total_inc_vat, created, modified
	`

	o := OrderRow{}
	currency := "GBP" // hardcoded for now but may come from elsewhere later.
	totalExVAT, totalVAT := totalSpend(cartProducts)
	totalIncVAT := totalExVAT + totalVAT

	row = tx.QueryRowContext(ctx, q6, c.id,
		bv.id, sv.id, currency, totalExVAT, totalVAT, totalIncVAT)
	err = row.Scan(&o.id, &o.UUID, &o.usrID, &o.Status, &o.Payment,
		&o.ContactName, &o.Email, &o.StripePI,
		&o.billingID, &o.shippingID, &o.Currency,
		&o.TotalExVAT, &o.VATTotal,
		&o.TotalIncVAT, &o.Created, &o.Modified)
	if err != nil {
		tx.Rollback()
		return nil, nil, nil, nil, nil, errors.Wrapf(err,
			"postgres: tx.QueryRowContext(ctx, q6=%q) failed", q6)
	}

	// 7. Insert the order items
	q7 := `
		INSERT INTO order_item (
		  order_id, path, sku, name,
		  qty, unit_price, discount,
		  tax_code, vat, created
		) VALUES (
		  $1, $2, $3, $4, $5, $6, $7, $8, $9, NOW()
		) RETURNING
		  id, uuid, order_id, path, sku, name, qty, unit_price, currency,
		  discount, tax_code, vat, created
	`
	stmt7, err := tx.PrepareContext(ctx, q7)
	if err != nil {
		tx.Rollback()
		return nil, nil, nil, nil, nil,
			errors.Wrapf(err, "postgres: tx prepare for q7=%q", q7)
	}
	defer stmt7.Close()

	orderItems := make([]*OrderItemRow, 0, len(cartProducts))
	for _, t := range cartProducts {
		oi := OrderItemRow{}
		row := stmt7.QueryRowContext(ctx, o.id,
			t.Path, t.SKU, t.Name,
			t.Qty, t.UnitPrice, nil,
			"T20", vat20Normalised(t.Qty*t.UnitPrice))
		err := row.Scan(&oi.id, &oi.UUID, &oi.orderID, &oi.Path, &oi.SKU,
			&oi.Name, &oi.Qty, &oi.UnitPrice, &oi.Currency,
			&oi.Discount, &oi.TaxCode, &oi.VAT, &oi.Created)
		if err != nil {
			tx.Rollback()
			return nil, nil, nil, nil, nil,
				errors.Wrap(err, "postgres: stmt.QueryRowContext failed")
		}
		orderItems = append(orderItems, &oi)
	}
	if err := tx.Commit(); err != nil {
		return nil, nil, nil, nil, nil,
			errors.Wrap(err, "postgres: tx.Commit() failed")
	}

	return &o, orderItems, &c, &bv, &sv, nil
}

// GetOrderDetailsByUUID retrieves the order row and order item rows
// for a given order. If the order cannot be found GetOrderDetailsByUUID
// returns an error == ErrOrderNotFound.
func (m *PgModel) GetOrderDetailsByUUID(ctx context.Context, orderUUID string) (*OrderRow, []*OrderItemRow, *OrderAddressRow, *OrderAddressRow, error) {
	contextLogger := log.WithContext(ctx)
	contextLogger.Debugf("postgres: GetOrderDetailsByUUID(ctx, orderUUID=%q", orderUUID)

	tx, err := m.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, nil, nil, nil,
			errors.Wrap(err, "postgres: db.BeginTx")
	}

	// 1. Get the main order details.
	q1 := `
		SELECT
		  id, uuid,
		  usr_id, status, payment, contact_name, email, stripe_pi,
		  billing_id, shipping_id, currency, total_ex_vat, vat_total,
		  total_inc_vat, created, modified
		FROM "order"
		WHERE uuid = $1
	`
	o := OrderRow{}
	err = tx.QueryRowContext(ctx, q1, orderUUID).Scan(&o.id, &o.UUID,
		&o.usrID, &o.Status, &o.Payment, &o.ContactName, &o.Email, &o.StripePI,
		&o.billingID, &o.shippingID, &o.Currency, &o.TotalExVAT, &o.VATTotal,
		&o.TotalIncVAT, &o.Created, &o.Modified)
	if err == sql.ErrNoRows {
		tx.Rollback()
		return nil, nil, nil, nil, ErrOrderNotFound
	}
	if err != nil {
		tx.Rollback()
		return nil, nil, nil, nil, errors.Wrapf(err,
			"postgres: query row context q1=%q", q1)
	}
	contextLogger.Debugf("postgres: q1 retrieved orderID=%d orderUUID=%q",
		o.id, orderUUID)

	// 2. Get the order product items.
	q2 := `
		SELECT
		  id, uuid, order_id, sku, name, qty, unit_price, currency, discount, tax_code, vat, created
		FROM order_item
		WHERE order_id = $1
	`
	rows, err := tx.QueryContext(ctx, q2, o.id)
	if err == sql.ErrNoRows {
		tx.Rollback()
		return nil, nil, nil, nil, ErrOrderItemsNotFound
	}
	if err != nil {
		tx.Rollback()
		return nil, nil, nil, nil, errors.Wrapf(err,
			"postgres: tx.QueryContext(ctx, q2=%q, order_id=%d) failed", q2, o.id)
	}
	defer rows.Close()

	orderProducts := make([]*OrderItemRow, 0, 16)
	for rows.Next() {
		i := OrderItemRow{}
		if err = rows.Scan(&i.id, &i.UUID, &i.orderID, &i.SKU, &i.Name, &i.Qty, &i.UnitPrice, &i.Currency, &i.Discount, &i.TaxCode, &i.VAT, &i.Created); err != nil {
			return nil, nil, nil, nil, errors.Wrapf(err, "postgres: scan order item %v", i)
		}
		orderProducts = append(orderProducts, &i)
	}
	if err := rows.Err(); err != nil {
		tx.Rollback()
		return nil, nil, nil, nil, errors.Wrapf(err, "postgres: rows err")
	}
	contextLogger.Debugf("postgres: q2 retrieved %d order items", len(orderProducts))

	// 3. Get the billing and shipping addresses.
	q3 := `
		SELECT
		  id, uuid, typ, contact_name, addr1, addr2, city, county,
		  postcode, country_code, created, modified
		FROM order_address
		WHERE id = $1
	`
	stmt3, err := tx.PrepareContext(ctx, q3)
	if err != nil {
		tx.Rollback()
		return nil, nil, nil, nil, errors.Wrapf(err,
			"postgres: tx prepare for q3=%q", q3)
	}
	defer stmt3.Close()

	var bv OrderAddressRow
	err = stmt3.QueryRowContext(ctx, o.shippingID).Scan(&bv.id, &bv.UUID,
		&bv.Typ, &bv.ContactName, &bv.Addr1, &bv.Addr2, &bv.City,
		&bv.County, &bv.Postcode, &bv.CountryCode,
		&bv.Created, &bv.Modified)
	if err != nil {
		tx.Rollback()
		return nil, nil, nil, nil,
			errors.Wrap(err, "postgres: scan failed")
	}

	var sv OrderAddressRow
	err = stmt3.QueryRowContext(ctx, o.shippingID).Scan(&sv.id, &sv.UUID,
		&sv.Typ, &sv.ContactName, &sv.Addr1, &sv.Addr2, &sv.City,
		&sv.County, &sv.Postcode, &sv.CountryCode,
		&sv.Created, &sv.Modified)
	if err != nil {
		tx.Rollback()
		return nil, nil, nil, nil,
			errors.Wrap(err, "postgres: scan failed")
	}

	if err := tx.Commit(); err != nil {
		tx.Rollback()
		return nil, nil, nil, nil,
			errors.Wrap(err, "postgres: tx.Commit() failed")
	}
	return &o, orderProducts, &bv, &sv, nil
}

// SetStripePaymentIntent sets payment intent id reference on an
// existing order and updates the modified timestamp.
func (m *PgModel) SetStripePaymentIntent(ctx context.Context, orderID, pi string) error {
	query := `
		UPDATE "order"
		SET stripe_pi = $1, modified = NOW()
		WHERE uuid = $2
	`
	_, err := m.db.ExecContext(ctx, query, pi, orderID)
	if err != nil {
		return errors.Wrapf(err,
			"postgres: m.db.ExecContext(ctx, query=%q, pi=%q, orderID=%q) failed",
			query, pi, orderID)
	}
	return nil
}

// RecordPayment marks the order with the given order ID and Stripe Intent
// referenceas complete and paid.
func (m *PgModel) RecordPayment(ctx context.Context, orderID, pi string, body []byte) error {
	contextLogger := log.WithContext(ctx)
	contextLogger.Infof("postgres: RecordPayment(ctx, orderID=%q, pi=%q, body=%v",
		orderID, pi, string(body))

	tx, err := m.db.BeginTx(ctx, nil)
	if err != nil {
		return errors.Wrap(err, "postgres: db.BeginTx")
	}

	query := `
		UPDATE "order"
		SET status = 'completed', payment = 'paid', modified = NOW()
		WHERE uuid = $1 AND stripe_pi = $2
		RETURNING id
	`
	var id int

	err = tx.QueryRowContext(ctx, query, orderID, pi).Scan(&id)
	if err != nil {
		return errors.Wrapf(err, "postgres: tx.ExecContext(ctx, query=%q, pi=%q, orderID=%q) failed", query, pi, orderID)
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
		return errors.Wrapf(err, "postgres: tx.ExecContext(ctx, query=%q, id=%d, body=%q) failed", query, id, body)
	}
	if err := tx.Commit(); err != nil {
		tx.Rollback()
		return errors.Wrap(err, "postgres: tx.Commit() failed")
	}
	return nil
}
