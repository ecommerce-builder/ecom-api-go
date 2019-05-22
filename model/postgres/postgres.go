package postgres

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"bitbucket.org/andyfusniakteam/ecom-api-go/utils/nestedset"
	"github.com/lib/pq"
	"github.com/pkg/errors"
	"golang.org/x/crypto/bcrypt"
)

// PgModel contains the database handle
type PgModel struct {
	db *sql.DB
}

// CartItem structure holds the details individual cart item.
type CartItem struct {
	ID        int
	CartUUID  string
	Sku       string
	Qty       int
	UnitPrice float64
	Created   time.Time
	Modified  time.Time
}

// Customer details.
type Customer struct {
	ID        int
	UUID      string
	UID       string
	Role      string
	Email     string
	Firstname string
	Lastname  string
	Created   time.Time
	Modified  time.Time
}

// CustomerDevKey customer developer keys.
type CustomerDevKey struct {
	ID           int
	UUID         string
	Key          string
	Hash         string
	CustomerID   int
	CustomerUUID string
	Created      time.Time
	Modified     time.Time
}

type Product struct {
	ID       int
	UUID     string
	SKU      string
	EAN      string
	URL      string
	Name     string
	Data     ProductData
	Created  time.Time
	Modified time.Time
}

type ProductUpdate struct {
	EAN  string      `json:"ean"`
	URL  string      `json:"url"`
	Name string      `json:"name"`
	Data ProductData `json:"data"`
}

type ProductData struct {
	Summary string `json:"summary"`
	Desc    string `json:"description"`
	Spec    string `json:"specification"`
}

type ProductImage struct {
	ID        uint
	ProductID uint
	UUID      string
	SKU       string
	W         uint
	H         uint
	Path      string
	Typ       string
	Ori       bool
	Up        bool
	Pri       uint
	Size      uint
	Q         uint
	GSURL     string
	Data      interface{}
	Created   time.Time
	Modified  time.Time
}

type CreateProductImage struct {
	SKU   string
	W     uint
	H     uint
	Path  string
	Typ   string
	Ori   bool
	Pri   uint
	Size  uint
	Q     uint
	GSURL string
	Data  interface{}
}

type CatalogProduct struct {
	ID        int
	CatalogID int
	ProductID int
	Path      string
	SKU       string
	Pri       int
	Created   time.Time
	Modified  time.Time
}

// Address contains address information for a Customer
type Address struct {
	ID          int
	AddrUUID    string
	CustomerID  int
	Typ         string
	ContactName string
	Addr1       string
	Addr2       *string
	City        string
	County      *string
	Postcode    string
	Country     string
	Created     time.Time
	Modified    time.Time
}

// PaginationResultSet contains both the underlying result set as well as
// context about the data including Total; the total number of rows in
// the table, First; set to true if this result set represents the first
// page, Last; set to true if this result set represents the last page of
// results.
type PaginationResultSet struct {
	RContext struct {
		Total               int
		FirstUUID, LastUUID string
	}
	RSet interface{}
}

type PaginationQuery struct {
	OrderBy    string
	OrderDir   string
	Limit      int
	StartAfter string
}

// CatalogProductAssoc maps products to leaf nodes in the catalogue hierarchy.
type CatalogProductAssoc struct {
	id        int
	catalogID int
	productID int
	Path      string
	SKU       string
	Pri       int
	Created   time.Time
	Modified  time.Time
}

func (pd ProductData) Value() (driver.Value, error) {
	bs, err := json.Marshal(pd)
	if err != nil {
		return nil, errors.Wrap(err, "json marshal failed")
	}
	return string(bs), nil
}

func (pd *ProductData) Scan(value interface{}) error {
	sv, err := driver.String.ConvertValue(value)
	if err != nil {
		return errors.Wrap(err, "convert value failed")
	}
	if v, ok := sv.([]byte); ok {
		var pdu ProductData
		err := json.Unmarshal(v, &pdu)
		if err != nil {
			return errors.Wrap(err, "json unmarshal failed")
		}
		*pd = pdu
		return nil
	}
	return fmt.Errorf("scan value failed")
}

// NewPgModel creates a new PgModel instance
func NewPgModel(db *sql.DB) *PgModel {
	return &PgModel{
		db: db,
	}
}

// IsAdmin returns true is the given customer UUID has a role of admin.
func (m *PgModel) IsAdmin(ctx context.Context, uuid string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM customers WHERE uuid=$1 AND role='admin') AS exists`
	var exists bool
	err := m.db.QueryRowContext(ctx, query, uuid).Scan(&exists)
	if err != nil {
		return false, errors.Wrapf(err, "db.QueryRow(ctx, %s)", query)
	}
	return exists, nil
}

// GetAdmin returns a Customer of role admin for the given UUID.
func (m *PgModel) GetAdmin(ctx context.Context, uuid string) (*Customer, error) {
	query := `
		SELECT
			id, uuid, uid, role, email, firstname, lastname, created, modified
		FROM customers
		WHERE uuid = $1 AND role='admin'
	`
	c := Customer{}
	err := m.db.QueryRowContext(ctx, query, uuid).Scan(&c.ID, &c.UUID, &c.UID, &c.Role, &c.Email, &c.Firstname, &c.Lastname, &c.Created, &c.Modified)
	if err != nil {
		return nil, errors.Wrapf(err, "query row context scan query=%q Customer=%v", query, c)
	}
	return &c, nil
}

// GetAllAdmins returns a slice of Customers who are all of role admin
func (m *PgModel) GetAllAdmins(ctx context.Context) ([]*Customer, error) {
	query := `
		SELECT id, uuid, uid, role, email, firstname, lastname, created, modified
		FROM customers
		WHERE role = 'admin'
		ORDER by created DESC
	`
	rows, err := m.db.QueryContext(ctx, query)
	if err != nil {
		return nil, errors.Wrapf(err, "m.db.QueryContext(ctx) query=%q", query)
	}
	defer rows.Close()

	admins := make([]*Customer, 0, 8)
	for rows.Next() {
		var c Customer
		err := rows.Scan(&c.ID, &c.UUID, &c.UID, &c.Role, &c.Email, &c.Firstname, &c.Lastname, &c.Created, &c.Modified)
		if err != nil {
			return nil, errors.Wrap(err, "scan failed")
		}
		admins = append(admins, &c)
	}
	if err = rows.Err(); err != nil {
		return nil, errors.Wrap(err, "rows.Err()")
	}
	return admins, nil
}

// DeleteAdminByUUID deletes the administrator from the customers table
// with the given UUID.
func (m *PgModel) DeleteAdminByUUID(ctx context.Context, uuid string) error {
	query := `DELETE FROM customers WHERE uuid = $1 AND role = 'admin'`
	_, err := m.db.ExecContext(ctx, query, uuid)
	if err != nil {
		return errors.Wrapf(err, "exec context query=%q", query)
	}
	return nil
}

// CreateCart creates a new shopping cart
func (m *PgModel) CreateCart(ctx context.Context) (*string, error) {
	var cartUUID string
	query := `SELECT UUID_GENERATE_V4() AS uuid`
	err := m.db.QueryRowContext(ctx, query).Scan(&cartUUID)
	if err != nil {
		return nil, errors.Wrapf(err, "db.QueryRow(ctx, %s)", query)
	}
	return &cartUUID, nil
}

// AddItemToCart adds a new item with sku, qty and unit price
func (m *PgModel) AddItemToCart(ctx context.Context, cartUUID, tierRef, sku string, qty int) (*CartItem, error) {
	item := CartItem{}

	// check if the item is already in the cart
	query := `SELECT EXISTS(SELECT 1 FROM carts WHERE uuid=$1 AND sku=$2) AS exists;`
	var exists bool
	m.db.QueryRowContext(ctx, query, cartUUID, sku).Scan(&exists)
	if exists == true {
		return nil, fmt.Errorf("cart item %s already exists", sku)
	}

	var unitPriceStr []byte
	query = `SELECT unit_price FROM product_pricing WHERE tier_ref = $1 AND sku = $2`
	err := m.db.QueryRowContext(ctx, query, tierRef, sku).Scan(&unitPriceStr)
	if err != nil {
		return &item, errors.Wrapf(err, "query scan failed query=%q", query)
	}

	unitPrice, _ := strconv.ParseFloat(string(unitPriceStr), 64)
	query = `
		INSERT INTO carts (uuid, sku, qty, unit_price)
		VALUES ($1, $2, $3, $4)
		RETURNING id, uuid, sku, qty, unit_price, created, modified
	`

	err = m.db.QueryRowContext(ctx, query, cartUUID, sku, qty, unitPrice).Scan(&item.ID, &item.CartUUID, &item.Sku, &item.Qty, &item.UnitPrice, &item.Created, &item.Modified)
	if err != nil {
		return nil, errors.Wrapf(err, "query scan failed query=%q", query)
	}
	return &item, nil
}

// GetCartItems gets all items in the cart
func (m *PgModel) GetCartItems(ctx context.Context, cartUUID string) ([]*CartItem, error) {
	query := `
		SELECT
			id, uuid, sku, qty, unit_price, created, modified
		FROM carts
		WHERE uuid = $1
	`
	rows, err := m.db.QueryContext(ctx, query, cartUUID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	cartItems := make([]*CartItem, 0, 20)
	for rows.Next() {
		c := CartItem{}
		err = rows.Scan(&c.ID, &c.CartUUID, &c.Sku, &c.Qty, &c.UnitPrice, &c.Created, &c.Modified)
		if err != nil {
			return nil, errors.Wrapf(err, "scan cart item %v", c)
		}
		cartItems = append(cartItems, &c)
	}

	if err = rows.Err(); err != nil {
		return nil, errors.Wrapf(err, "rows err")
	}
	return cartItems, nil
}

// UpdateItemByCartUUID updates the qty of a cart item of the given sku.
func (m *PgModel) UpdateItemByCartUUID(ctx context.Context, cartUUID, sku string, qty int) (*CartItem, error) {
	query := `
		UPDATE carts
		SET qty = $1, modified = NOW()
		WHERE uuid = $2 AND sku = $3
		RETURNING id, uuid, sku, qty, unit_price, created, modified
	`

	item := CartItem{}
	err := m.db.QueryRowContext(ctx, query, qty, cartUUID, sku).Scan(&item.ID, &item.CartUUID, &item.Sku, &item.Qty, &item.UnitPrice, &item.Created, &item.Modified)
	if err != nil {
		return nil, errors.Wrapf(err, "query row scan query=%q", query)
	}
	return &item, nil
}

// DeleteCartItem deletes a single cart item.
func (m *PgModel) DeleteCartItem(ctx context.Context, cartUUID, sku string) (count int64, err error) {
	query := `DELETE FROM carts WHERE uuid = $1 AND sku = $2`
	res, err := m.db.ExecContext(ctx, query, cartUUID, sku)
	if err != nil {
		return -1, errors.Wrapf(err, "exec context query=%q", query)
	}
	count, err = res.RowsAffected()
	if err != nil {
		return -1, errors.Wrap(err, "rows affected")
	}
	return count, nil
}

// EmptyCartItems empty the cart of all items. Does not affect coupons.
func (m *PgModel) EmptyCartItems(ctx context.Context, cartUUID string) (err error) {
	query := `DELETE FROM carts WHERE uuid = $1`
	_, err = m.db.ExecContext(ctx, query, cartUUID)
	if err != nil {
		return errors.Wrapf(err, "exec context query=%q", query)
	}
	return nil
}

// CreateCustomer creates a new customer
func (m *PgModel) CreateCustomer(ctx context.Context, uid, role, email, firstname, lastname string) (*Customer, error) {
	query := `
		INSERT INTO customers (
			uid, role, email, firstname, lastname
		) VALUES (
			$1, $2, $3, $4, $5
		)
		RETURNING id, uuid, uid, role, email, firstname, lastname, created, modified
	`
	c := Customer{}
	err := m.db.QueryRowContext(ctx, query, uid, role, email, firstname, lastname).Scan(
		&c.ID, &c.UUID, &c.UID, &c.Role, &c.Email, &c.Firstname, &c.Lastname, &c.Created, &c.Modified)
	if err != nil {
		return nil, errors.Wrapf(err, "query row context Customer=%v", c)
	}
	return &c, nil
}

// GetCustomers gets the next size customers starting at page page
func (m *PgModel) GetCustomers(ctx context.Context, pq *PaginationQuery) (*PaginationResultSet, error) {
	q := NewQuery("customers", map[string]bool{
		"id":        true,
		"uuid":      false,
		"uid":       false,
		"role":      true,
		"email":     true,
		"firstname": true,
		"lastname":  true,
		"created":   true,
		"modified":  true,
	})
	q = q.Select([]string{"id", "uuid", "uid", "role", "email", "firstname", "lastname", "created", "modified"})

	// if not set, default Order By, Order Direction and Limit is "created DESC LIMIT 10"
	if pq.OrderBy != "" {
		q = q.OrderBy(pq.OrderBy)
	} else {
		q = q.OrderBy("created")
	}
	if pq.OrderDir != "" {
		q = q.OrderDir(OrderDirection(pq.OrderDir))
	} else {
		q = q.OrderDir("DESC")
	}
	if pq.Limit > 0 {
		q = q.Limit(pq.Limit)
	} else {
		q = q.Limit(10)
	}
	if pq.StartAfter != "" {
		q = q.StartAfter(pq.StartAfter)
	}

	// calculate the total count, first and last items in the result set
	pr := PaginationResultSet{}
	sql := `
		SELECT COUNT(*) AS count
		FROM %s
	`
	sql = fmt.Sprintf(sql, q.table)
	err := m.db.QueryRowContext(ctx, sql).Scan(&pr.RContext.Total)
	if err != nil {
		return nil, errors.Wrapf(err, "query row context scan query=%q", sql)
	}

	// book mark either end of the result set
	sql = `
		SELECT uuid
		FROM %s
		ORDER BY %s %s, id %s
		FETCH FIRST 1 ROW ONLY
	`
	sql = fmt.Sprintf(sql, q.table, q.orderBy, string(q.orderDir), string(q.orderDir))
	err = m.db.QueryRowContext(ctx, sql).Scan(&pr.RContext.FirstUUID)
	if err != nil {
		return nil, errors.Wrapf(err, "query row context query=%q", sql)
	}
	sql = `
		SELECT uuid
		FROM %s
		ORDER BY %s %s, id %s
		FETCH FIRST 1 ROW ONLY
	`
	sql = fmt.Sprintf(sql, q.table, q.orderBy, string(q.orderDir.toggle()), string(q.orderDir.toggle()))
	err = m.db.QueryRowContext(ctx, sql).Scan(&pr.RContext.LastUUID)
	if err != nil {
		return nil, errors.Wrapf(err, "query row context query=%q", sql)
	}

	rows, err := m.QueryContextQ(ctx, q)
	if err != nil {
		return nil, errors.Wrapf(err, "model query context q=%v", q)
	}
	defer rows.Close()

	customers := make([]*Customer, 0)
	for rows.Next() {
		var c Customer
		err = rows.Scan(&c.ID, &c.UUID, &c.UID, &c.Role, &c.Email, &c.Firstname, &c.Lastname, &c.Created, &c.Modified)
		if err != nil {
			return nil, errors.Wrapf(err, "rows scan Customer=%v", c)
		}
		customers = append(customers, &c)
	}

	if err = rows.Err(); err != nil {
		return nil, errors.Wrap(err, "rows err")
	}
	pr.RSet = customers
	return &pr, nil
}

// GetCustomerByUUID gets a customer by customer UUID
func (m *PgModel) GetCustomerByUUID(ctx context.Context, customerUUID string) (*Customer, error) {
	query := `
		SELECT
			id, uuid, uid, role, email, firstname, lastname, created, modified
		FROM customers
		WHERE uuid = $1
	`
	c := Customer{}
	err := m.db.QueryRowContext(ctx, query, customerUUID).Scan(&c.ID, &c.UUID, &c.UID, &c.Role, &c.Email, &c.Firstname, &c.Lastname, &c.Created, &c.Modified)
	if err != nil {
		return nil, errors.Wrapf(err, "query row context scan query=%q Customer=%v", query, c)
	}
	return &c, nil
}

// GetCustomerByID gets a customer by customer ID
func (m *PgModel) GetCustomerByID(ctx context.Context, customerID int) (*Customer, error) {
	query := `
		SELECT
			id, uuid, uid, role, email, firstname, lastname, created, modified
		FROM customers
		WHERE id = $1
	`
	c := Customer{}
	err := m.db.QueryRowContext(ctx, query, customerID).Scan(&c.ID, &c.UUID, &c.UID, &c.Role, &c.Email, &c.Firstname, &c.Lastname, &c.Created, &c.Modified)
	if err != nil {
		return nil, errors.Wrapf(err, "query row context scan query=%q Customer=%v", query, c)
	}
	return &c, nil
}

// CreateAddress creates a new billing or shipping address for a customer
func (m *PgModel) CreateAddress(ctx context.Context, customerID int, typ, contactName, addr1 string, addr2 *string, city string, county *string, postcode, country string) (*Address, error) {
	a := Address{}
	query := `
		INSERT INTO addresses (
			customer_id, typ, contact_name, addr1, addr2, city, county, postcode, country
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9
		) RETURNING
			id, uuid, customer_id, typ, contact_name, addr1, addr2, city, county, postcode, country, created, modified
	`
	err := m.db.QueryRowContext(ctx, query, customerID, typ, contactName, addr1, addr2, city, county, postcode, country).Scan(&a.ID, &a.AddrUUID, &a.CustomerID, &a.Typ, &a.ContactName, &a.Addr1, &a.Addr2, &a.City, &a.County, &a.Postcode, &a.Country, &a.Created, &a.Modified)
	if err != nil {
		return nil, errors.Wrapf(err, "query row context scan query=%q", query)
	}
	return &a, nil
}

// CreateProduct creates a new product with the given SKU.
func (m *PgModel) CreateProduct(ctx context.Context, sku string, pu *ProductUpdate) (*Product, error) {
	query := `
		INSERT INTO products (
			sku, ean, url, name, data, created, modified
		) VALUES (
			$1, $2, $3, $4, $5, NOW(), NOW()
		) RETURNING
			id, uuid, sku, ean, url, name, data, created, modified
	`
	p := Product{}
	err := m.db.QueryRowContext(ctx, query, sku, pu.EAN, pu.URL, pu.Name, pu.Data).Scan(&p.ID, &p.UUID, &p.SKU, &p.EAN, &p.URL, &p.Name, &p.Data, &p.Created, &p.Modified)
	if err != nil {
		return nil, errors.Wrapf(err, "query scan context sku=%q, query=%q", sku, query)
	}
	return &p, nil
}

// GetProduct returns the product by SKU.
func (m *PgModel) GetProduct(ctx context.Context, sku string) (*Product, error) {
	query := `
		SELECT id, uuid, sku, ean, url, name, data, created, modified
		FROM products
		WHERE sku = $1
	`
	p := Product{}
	err := m.db.QueryRowContext(ctx, query, sku).Scan(&p.ID, &p.UUID, &p.SKU, &p.EAN, &p.URL, &p.Name, &p.Data, &p.Created, &p.Modified)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, err
		}
		return nil, errors.Wrapf(err, "query scan context sku=%q query=%q", sku, query)
	}
	return &p, nil
}

// GetProducts returns a list of all products in the products table.
func (m *PgModel) GetProducts(ctx context.Context) ([]*Product, error) {
	query := `
		SELECT id, uuid, sku, ean, url, name, data, created, modified
		FROM products
	`
	rows, err := m.db.QueryContext(ctx, query)
	if err != nil {
		return nil, errors.Wrapf(err, "m.db.QueryContext(ctx) query=%q", query)
	}
	defer rows.Close()

	products := make([]*Product, 0, 256)
	for rows.Next() {
		var p Product
		err := rows.Scan(&p.ID, &p.UUID, &p.SKU, &p.EAN, &p.URL, &p.Name, &p.Data, &p.Created, &p.Modified)
		if err != nil {
			return nil, errors.Wrap(err, "scan failed")
		}
		products = append(products, &p)
	}
	if err = rows.Err(); err != nil {
		return nil, errors.Wrap(err, "rows.Err()")
	}
	return products, nil
}

// ProductsExist accepts a slice of product SKU strings and returns only
// those that can be found in the products table.
func (m *PgModel) ProductsExist(ctx context.Context, skus []string) ([]string, error) {
	query := `
		SELECT sku FROM products
		WHERE sku = ANY($1::varchar[])
	`
	// TODO: sanitise skus
	rows, err := m.db.QueryContext(ctx, query, "{"+strings.Join(skus, ",")+"}")
	if err != nil {
		return nil, errors.Wrapf(err, "m.db.QueryContext(ctx,..) query=%q, skus=%v", query, skus)
	}
	defer rows.Close()

	found := make([]string, 0, 256)
	for rows.Next() {
		var s string
		err = rows.Scan(&s)
		if err != nil {
			return nil, errors.Wrap(err, "scan failed")
		}
		found = append(found, s)
	}
	if err = rows.Err(); err != nil {
		return nil, errors.Wrap(err, "rows.Err()")
	}
	return found, nil
}

// ProductExists return true if there is a row in the products table with
// the given SKU.
func (m *PgModel) ProductExists(ctx context.Context, sku string) (bool, error) {
	query := `SELECT id FROM products WHERE sku = $1`
	var id int
	err := m.db.QueryRowContext(ctx, query, sku).Scan(&id)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		return false, errors.Wrapf(err, "query row context sku=%q query=%q", sku, query)
	}
	return true, nil
}

// UpdateProduct updates the details of a product with the given SKU.
func (m *PgModel) UpdateProduct(ctx context.Context, sku string, pu *ProductUpdate) (*Product, error) {
	query := `
		UPDATE products
		SET ean = $1, url = $2, name = $3, data = $4, modified = NOW()
		WHERE sku = $5
		RETURNING
			id, uuid, sku, ean, url, name, data, created, modified
	`
	p := Product{}
	err := m.db.QueryRowContext(ctx, query, pu.EAN, pu.URL, pu.Name, pu.Data, sku).Scan(
		&p.ID, &p.UUID, &p.SKU, &p.EAN, &p.URL, &p.Name, &p.Data, &p.Created, &p.Modified)
	if err != nil {
		return nil, errors.Wrapf(err, "query row context query=%q", query)
	}
	return &p, nil
}

// DeleteProduct delete the product with the given SKU returning the nu
func (m *PgModel) DeleteProduct(ctx context.Context, sku string) error {
	query := `
		DELETE FROM products
		WHERE sku = $1
	`
	_, err := m.db.ExecContext(ctx, query, sku)
	if err != nil {
		return errors.Wrapf(err, "exec context query=%q", query)
	}
	return nil
}

// CreateCustomerDevKey generates a customer developer key using bcrypt.
func (m *PgModel) CreateCustomerDevKey(ctx context.Context, customerID int, key string) (*CustomerDevKey, error) {
	query := `
		INSERT INTO customers_devkeys (
			key, hash, customer_id, created, modified
		) VALUES (
			$1, $2, $3, NOW(), NOW()
		) RETURNING
			id, key, hash, customer_id, created, modified
	`
	cdk := CustomerDevKey{}
	hash, err := bcrypt.GenerateFromPassword([]byte(key), 14)
	err = m.db.QueryRowContext(ctx, query, key, string(hash), customerID).Scan(&cdk.ID, &cdk.Key, &cdk.Hash, &cdk.CustomerID, &cdk.Created, &cdk.Modified)
	if err != nil {
		return nil, errors.Wrapf(err, "query row context scan query=%q", query)
	}
	return &cdk, nil
}

// GetCustomerDevKeys returns a slice of CustomerDevKeys by customer primary key.
func (m *PgModel) GetCustomerDevKeys(ctx context.Context, customerID int) ([]*CustomerDevKey, error) {
	query := `
		SELECT id, uuid, key, hash, customer_id, created, modified
		FROM customers_devkeys
		WHERE customer_id = $1
	`
	rows, err := m.db.QueryContext(ctx, query, customerID)
	if err != nil {
		return nil, errors.Wrapf(err, "m.db.QueryContext(ctx, %q, %d)", query, customerID)
	}
	defer rows.Close()

	apiKeys := make([]*CustomerDevKey, 0, 8)
	for rows.Next() {
		var c CustomerDevKey
		err = rows.Scan(&c.ID, &c.UUID, &c.Key, &c.Hash, &c.CustomerID, &c.Created, &c.Modified)
		if err != nil {
			return nil, errors.Wrap(err, "Scan failed")
		}
		apiKeys = append(apiKeys, &c)
	}
	if err = rows.Err(); err != nil {
		return nil, errors.Wrap(err, "rows.Err()")
	}
	return apiKeys, nil
}

// GetCustomerDevKey
func (m *PgModel) GetCustomerDevKey(ctx context.Context, uuid string) (*CustomerDevKey, error) {
	query := `
		SELECT A.id, A.uuid, key, hash, customer_id, C.uuid, A.created, A.modified
		FROM customers_devkeys AS A
		INNER JOIN customers AS C ON A.customer_id = C.id
		WHERE
			  A.uuid = $1
	`
	cak := CustomerDevKey{}
	err := m.db.QueryRowContext(ctx, query, uuid).Scan(&cak.ID, &cak.UUID, &cak.Key, &cak.Hash, &cak.CustomerID, &cak.CustomerUUID, &cak.Created, &cak.Modified)
	if err != nil {
		return nil, errors.Wrapf(err, " m.db.QueryRowContext(ctx, %q, %q).Scan(...)", query, uuid)
	}
	return &cak, nil
}

// GetCustomerDevKeyByDevKey retrieves a given Developer Key record.
func (m *PgModel) GetCustomerDevKeyByDevKey(ctx context.Context, key string) (*CustomerDevKey, error) {
	query := `
		SELECT
			A.id as id, A.uuid as uuid, key, hash, customer_id,
			C.uuid as customer_uuid, A.created as created, A.modified as modified
		FROM customers_devkeys AS A
		INNER JOIN customers AS C ON A.customer_id = C.id
		WHERE key = $1
	`
	cak := CustomerDevKey{}
	err := m.db.QueryRowContext(ctx, query, key).Scan(&cak.ID, &cak.UUID, &cak.Key, &cak.Hash, &cak.CustomerID, &cak.CustomerUUID, &cak.Created, &cak.Modified)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, err
		}
		return nil, errors.Wrapf(err, "m.db.QueryRowContext(ctx, %q, %q).Scan(...)", query, key)
	}
	return &cak, nil
}

// GetAddressByUUID gets an address by UUID. Returns a pointer to an Address.
func (m *PgModel) GetAddressByUUID(ctx context.Context, uuid string) (*Address, error) {
	a := Address{}
	query := `
		SELECT
			id, uuid, customer_id, typ, contact_name, addr1, addr2,
			city, county, postcode, country, created, modified
		FROM addresses
		WHERE uuid = $1
	`
	err := m.db.QueryRowContext(ctx, query, uuid).Scan(&a.ID, &a.AddrUUID, &a.CustomerID, &a.Typ, &a.ContactName, &a.Addr1, &a.Addr2, &a.City, &a.County, &a.Postcode, &a.Country, &a.Created, &a.Modified)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, &ResourceError{
				Op:       "GetAddressByUUID",
				Resource: "address",
				UUID:     uuid,
				Err:      ErrNotExist,
			}
		}
		if pge, ok := err.(*pq.Error); ok {
			switch pge.Code.Name() {
			case "invalid_text_representation":
				return nil, &ResourceError{
					Op:       "GetAddressByUUID",
					Resource: "address",
					UUID:     uuid,
					Err:      ErrInvalidText,
				}
			default:
				return nil, pge
			}
		}
		return nil, errors.Wrapf(err, "query row context scan query=%q", query)
	}
	return &a, nil
}

// GetAddressOwnerByUUID returns a pointer to a string containing the
// customer UUID of the owner of this address record. If the address is not
// found the return value of will be nil.
func (m *PgModel) GetAddressOwnerByUUID(ctx context.Context, uuid string) (*string, error) {
	query := `
		SELECT C.uuid
		FROM customers AS C, addresses AS A
		WHERE A.customer_id = C.id AND A.uuid = $1
	`
	var customerUUID string
	err := m.db.QueryRowContext(ctx, query, uuid).Scan(&customerUUID)
	if err != nil {
		return nil, errors.Wrapf(err, "query row context query=%q", query)
	}
	return &customerUUID, nil
}

// GetCustomerIDByUUID converts between customer UUID and the underlying
// primary key.
func (m *PgModel) GetCustomerIDByUUID(ctx context.Context, customerUUID string) (int, error) {
	var id int
	query := `SELECT id FROM customers WHERE uuid = $1`
	row := m.db.QueryRowContext(ctx, query, customerUUID)
	err := row.Scan(&id)
	if err != nil {
		return -1, errors.Wrapf(err, "query row context query=%q", query)
	}
	return id, nil
}

// GetAddresses retrieves a slice of pointers to Address for a given
// customer.
func (m *PgModel) GetAddresses(ctx context.Context, customerID int) ([]*Address, error) {
	addresses := make([]*Address, 0, 8)
	query := `
		SELECT
			id, uuid, customer_id, typ, contact_name, addr1,
			addr2, city, county, postcode, country, created, modified
		FROM addresses
		WHERE customer_id = $1
		ORDER BY created DESC
	`
	rows, err := m.db.QueryContext(ctx, query, customerID)
	if err != nil {
		return nil, errors.Wrapf(err, "db query context query=%q", query)
	}
	defer rows.Close()

	for rows.Next() {
		var a Address
		err = rows.Scan(&a.ID, &a.AddrUUID, &a.CustomerID, &a.Typ, &a.ContactName, &a.Addr1, &a.Addr2, &a.City, &a.County, &a.Postcode, &a.Country, &a.Created, &a.Modified)
		if err != nil {
			return nil, errors.Wrapf(err, "rows scan query=%q", query)
		}
		addresses = append(addresses, &a)
	}

	if err = rows.Err(); err != nil {
		return nil, errors.Wrap(err, "rows err")
	}
	return addresses, nil
}

// UpdateAddressByUUID updates an address for a given customer
func (m *PgModel) UpdateAddressByUUID(ctx context.Context, addrUUID string) (*Address, error) {
	// TO BE DONE
	//
	//query := `UPDATE addresses SET`
	addr := Address{}
	return &addr, nil
}

// DeleteAddressByUUID deletes an address by uuid
func (m *PgModel) DeleteAddressByUUID(ctx context.Context, addrUUID string) error {
	query := `DELETE FROM addresses WHERE uuid = $1`
	_, err := m.db.ExecContext(ctx, query, addrUUID)
	if err != nil {
		return errors.Wrapf(err, "exec context query=%q", query)
	}
	return nil
}

// BatchCreateNestedSet creates a nested set of nodes representing the
// catalog hierarchy.
func (m *PgModel) BatchCreateNestedSet(ctx context.Context, ns []*nestedset.NestedSetNode) error {
	tx, err := m.db.BeginTx(ctx, nil)
	if err != nil {
		return errors.Wrap(err, "db.BeginTx")
	}
	query := "DELETE FROM catalog"
	if _, err = tx.ExecContext(ctx, query); err != nil {
		tx.Rollback()
		return errors.Wrapf(err, "model: delete catalog query=%q", query)
	}
	query = `
		INSERT INTO catalog (
			segment, path, name, lft, rgt, depth, created, modified
		) VALUES (
			$1, $2, $3, $4, $5, $6, NOW(), NOW()
		)
	`
	stmt, err := tx.PrepareContext(ctx, query)
	if err != nil {
		tx.Rollback()
		return errors.Wrapf(err, "tx prepare for query=%q", query)
	}
	defer stmt.Close()
	for _, n := range ns {
		if _, err := stmt.ExecContext(ctx, n.Segment, n.Path, n.Name, n.Lft, n.Rgt, n.Depth); err != nil {
			tx.Rollback() // return an error too, we may want to wrap them
			return errors.Wrapf(err, "stmt exec segment=%q path=%q name=%q lft=%d rgt=%d depth=%d", n.Segment, n.Path, n.Name, n.Lft, n.Rgt, n.Depth)
		}
	}
	if err = tx.Commit(); err != nil {
		return errors.Wrap(err, "tx.Commit")
	}
	return nil
}

// GetCatalogByPath retrieves a single set element by the given path.
func (m *PgModel) GetCatalogByPath(ctx context.Context, path string) (*nestedset.NestedSetNode, error) {
	query := `
		SELECT id, segment, path, name, lft, rgt, depth, created, modified
		FROM catalog
		WHERE path = $1
	`
	var n nestedset.NestedSetNode
	err := m.db.QueryRowContext(ctx, query, path).Scan(&n.ID, &n.Segment, &n.Path, &n.Name, &n.Lft, &n.Rgt, &n.Depth, &n.Created, &n.Modified)
	if err != nil {
		return nil, errors.Wrapf(err, "service: query row ctx scan query=%q", query)
	}
	return &n, nil
}

// HasCatalog returns true if any rows exist in the catalog table.
func (m *PgModel) HasCatalog(ctx context.Context) (bool, error) {
	query := "SELECT COUNT(*) AS count FROM catalog"
	var count int
	err := m.db.QueryRowContext(ctx, query).Scan(&count)
	if err != nil {
		return false, errors.Wrapf(err, "query row context scan query=%q", query)
	}
	if count == 0 {
		return false, nil
	}
	return true, nil
}

// GetCatalogNestedSet returns a slice of NestedSetNode representing the catalog as a nested set.
func (m *PgModel) GetCatalogNestedSet(ctx context.Context) ([]*nestedset.NestedSetNode, error) {
	query := `
		SELECT id, segment, path, name, lft, rgt, depth, created, modified
		FROM catalog
		ORDER BY lft ASC
	`
	rows, err := m.db.QueryContext(ctx, query)
	if err != nil {
		return nil, errors.Wrapf(err, "query context query=%q", query)
	}
	defer rows.Close()

	nodes := make([]*nestedset.NestedSetNode, 0, 256)
	for rows.Next() {
		var n nestedset.NestedSetNode
		err = rows.Scan(&n.ID, &n.Segment, &n.Path, &n.Name, &n.Lft, &n.Rgt, &n.Depth, &n.Created, &n.Modified)
		if err != nil {
			return nil, err
		}
		nodes = append(nodes, &n)
	}

	if err = rows.Err(); err != nil {
		return nil, errors.Wrap(err, "rows err")
	}
	return nodes, nil
}

// DeleteCatalogNestedSet delete all rows in the catalog table.
func (m *PgModel) DeleteCatalogNestedSet(ctx context.Context) error {
	query := `DELETE FROM catalog`
	_, err := m.db.ExecContext(ctx, query)
	if err != nil {
		return errors.Wrap(err, "service: delete catalog")
	}
	return nil
}

// CreateCatalogProductAssoc links an existing product identified by sku
// to an existing leaf node of the catalog denoted by path.
func (m *PgModel) CreateCatalogProductAssoc(ctx context.Context, path, sku string) (*CatalogProduct, error) {
	query := `
		INSERT INTO catalog_products
			(catalog_id, product_id, path, sku, pri)
		VALUES (
			(SELECT id FROM catalog WHERE path = $1),
			(SELECT id FROM products WHERE sku = $2),
			$3,
			$4,
			(
				SELECT
					CASE WHEN COUNT(1) > 0
					THEN MAX(pri)+10
					ELSE 10
				END
				AS pri
				FROM catalog_products
				WHERE path=$5
			)
		)
		RETURNING
			id, catalog_id, product_id, path, sku, pri, created, modified
	`
	cp := CatalogProduct{}
	err := m.db.QueryRowContext(ctx, query, path, sku, path, sku, path).Scan(&cp.ID, &cp.CatalogID, &cp.ProductID, &cp.Path, &cp.SKU, &cp.Pri, &cp.Created, &cp.Modified)
	if err != nil {
		return nil, errors.Wrapf(err, "model: query row context scan query=%q", query)
	}
	return &cp, nil
}

// BatchCreateCatalogProductAssocs inserts multiple catalog product
// associations using a transaction.
func (m *PgModel) BatchCreateCatalogProductAssocs(ctx context.Context, cpas map[string][]string) error {
	tx, err := m.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	query := "DELETE FROM catalog_products"
	_, err = tx.ExecContext(ctx, query)
	if err != nil {
		tx.Rollback()
		return errors.Wrapf(err, "model: delete catalog_products query=%q", query)
	}

	query = `
		INSERT INTO catalog_products
			(catalog_id, product_id, path, sku, pri)
		VALUES (
			(SELECT id FROM catalog WHERE path = $1),
			(SELECT id FROM products WHERE sku = $2),
			$3,
			$4,
			(
				SELECT
					CASE WHEN COUNT(1) > 0
					THEN MAX(pri)+10
					ELSE 10
				END
				AS pri
				FROM catalog_products
				WHERE path=$5
			)
		)
	`
	stmt, err := tx.PrepareContext(ctx, query)
	if err != nil {
		tx.Rollback()
		return errors.Wrapf(err, "tx prepare for query=%q", query)
	}
	defer stmt.Close()

	for path, skus := range cpas {
		for _, sku := range skus {
			if _, err := stmt.ExecContext(ctx, path, sku, path, sku, path); err != nil {
				tx.Rollback()
				fmt.Fprintf(os.Stderr, "%+v", err)
				return errors.Wrap(err, "stmt exec context")
			}
		}
	}
	return tx.Commit()
}

// DeleteCatalogProductAssoc delete an existing catalog product association.
func (m *PgModel) DeleteCatalogProductAssoc(ctx context.Context, path, sku string) error {
	query := `
		DELETE FROM catalog_products
		WHERE path = $1 AND sku = $2
	`
	_, err := m.db.ExecContext(ctx, query, path, sku)
	if err != nil {
		return errors.Wrapf(err, "service: delete catalog product assoc path=%q sku=%q", path, sku)
	}
	return nil
}

// HasCatalogProductAssocs returns true if any catalog product associations
// exist.
func (m *PgModel) HasCatalogProductAssocs(ctx context.Context) (bool, error) {
	query := "SELECT COUNT(*) AS count FROM catalog_products"
	var count int
	err := m.db.QueryRowContext(ctx, query).Scan(&count)
	if err != nil {
		return false, errors.Wrapf(err, "query row context scan query=%q", query)
	}
	if count == 0 {
		return false, nil
	}
	return true, nil
}

// GetCatalogProductAssocs returns an Slice of catalogue to product
// associations.
func (m *PgModel) GetCatalogProductAssocs(ctx context.Context) ([]*CatalogProductAssoc, error) {
	query := `
		SELECT id, catalog_id, product_id, path, sku, pri, created, modified
		FROM catalog_products
		ORDER BY path, pri ASC
	`
	rows, err := m.db.QueryContext(ctx, query)
	if err != nil {
		return nil, errors.Wrapf(err, "model: query context query=%q", query)
	}
	defer rows.Close()

	cpas := make([]*CatalogProductAssoc, 0, 256)
	for rows.Next() {
		var n CatalogProductAssoc
		err = rows.Scan(&n.id, &n.catalogID, &n.productID, &n.Path, &n.SKU, &n.Pri, &n.Created, &n.Modified)
		if err != nil {
			return nil, errors.Wrapf(err, "model: scan failed")
		}
		cpas = append(cpas, &n)
	}
	if err = rows.Err(); err != nil {
		return nil, errors.Wrapf(err, "model: rows.Err()")
	}
	return cpas, nil
}

// UpdateCatalogProductAssocs update the catalog product associations.
func (m *PgModel) UpdateCatalogProductAssocs(ctx context.Context, cpo []*CatalogProductAssoc) error {
	tx, err := m.db.Begin()
	if err != nil {
		return err
	}

	stmt, err := tx.Prepare(`
		INSERT INTO catalog_products
			(catalog_id, product_id, path, sku, pri)
		VALUES (
			(SELECT id FROM catalog WHERE path = $1),
			(SELECT id FROM products WHERE sku = $2),
			$3,
			$4,
			$5
		)
	`)
	if err != nil {
		tx.Rollback()
		fmt.Fprintf(os.Stderr, "%v", err)
		return err
	}
	defer stmt.Close()

	for _, c := range cpo {
		if _, err := stmt.ExecContext(ctx, c.Path, c.SKU, c.Path, c.SKU, c.Pri); err != nil {
			tx.Rollback()
			fmt.Fprintf(os.Stderr, "%v", err)
			return err
		}
	}
	return tx.Commit()
}

// DeleteCatalogProductAssocs delete all catalog product associations.
func (m *PgModel) DeleteCatalogProductAssocs(ctx context.Context) (affected int64, err error) {
	res, err := m.db.ExecContext(ctx, "DELETE FROM catalog_products")
	if err != nil {
		return -1, errors.Wrap(err, "service: delete catalog product assocs")
	}
	count, err := res.RowsAffected()
	if err != nil {
		return -1, errors.Wrap(err, "service: rows affected")
	}
	return count, nil
}

// CreateImageEntry writes a new image entry to the product_images table.
func (m *PgModel) CreateImageEntry(ctx context.Context, c *CreateProductImage) (*ProductImage, error) {
	query := `
		INSERT INTO product_images (
			product_id, sku,
			w, h, path, typ,
			ori, up,
			pri, size, q,
			gsurl, created, modified
		) VALUES (
			(SELECT id FROM products WHERE sku = $1), $2,
			$3, $4, $5, $6,
			$7, false,
			$8, $9, $10,
			$11, NOW(), NOW()
		) RETURNING
			id, product_id, uuid, sku, w, h, path, typ, ori, up, pri, size, q,
			gsurl, data, created, modified
	`
	p := ProductImage{}
	err := m.db.QueryRowContext(ctx, query, c.SKU, c.SKU,
		c.W, c.H, c.Path, c.Typ,
		c.Ori,
		c.Pri, c.Size, c.Q,
		c.GSURL).Scan(&p.ID, &p.ProductID, &p.UUID, &p.SKU, &p.W, &p.H, &p.Path, &p.Typ, &p.Ori, &p.Up, &p.Pri, &p.Size, &p.Q, &p.GSURL, &p.Data, &p.Created, &p.Modified)
	if err != nil {
		return nil, err
	}

	return &p, nil
}

// GetImageEntries returns a list of all images associated to a given product SKU.
func (m *PgModel) GetImageEntries(ctx context.Context, sku string) ([]*ProductImage, error) {
	query := `
		SELECT id, product_id, uuid, sku, w, h, path, typ, ori, up, pri, size, q,
		gsurl, data, created, modified
		FROM product_images
		WHERE sku = $1
		ORDER BY pri ASC
	`
	rows, err := m.db.QueryContext(ctx, query, sku)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	images := make([]*ProductImage, 0, 16)
	for rows.Next() {
		p := ProductImage{}
		err = rows.Scan(&p.ID, &p.ProductID, &p.UUID, &p.SKU, &p.W, &p.H, &p.Path, &p.Typ, &p.Ori, &p.Up, &p.Pri, &p.Size, &p.Q, &p.GSURL, &p.Data, &p.Created, &p.Modified)
		if err != nil {
			return nil, err
		}
		images = append(images, &p)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}
	return images, nil
}

// ConfirmImageUploaded updates the `up` column to true to indicate the
// uploaded has taken place.
func (m *PgModel) ConfirmImageUploaded(ctx context.Context, uuid string) (*ProductImage, error) {
	query := `
		UPDATE product_images
		SET up = 't', modified = NOW()
		WHERE uuid = $1
		RETURNING id, product_id, uuid, sku, w, h, path, typ, ori, up, pri, size, q,
		gsurl, data, created, modified
	`
	p := ProductImage{}
	err := m.db.QueryRowContext(ctx, query, uuid).Scan(&p.ID, &p.ProductID, &p.UUID, &p.SKU, &p.W, &p.H, &p.Path, &p.Typ, &p.Ori, &p.Up, &p.Pri, &p.Size, &p.Q, &p.GSURL, &p.Data, &p.Created, &p.Modified)
	if err != nil {
		return nil, err
	}
	return &p, nil
}

// DeleteImageEntry deletes an image entry row from the product_images
// table by UUID.
func (m *PgModel) DeleteImageEntry(ctx context.Context, uuid string) (int64, error) {
	query := `
		DELETE FROM product_images
		WHERE uuid = $1
	`
	res, err := m.db.ExecContext(ctx, query, uuid)
	if err != nil {
		return -1, err
	}
	count, err := res.RowsAffected()
	if err != nil {
		return -1, err
	}
	return count, nil
}
