package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"strconv"

	"bitbucket.org/andyfusniakteam/ecom-api-go/model"
	"bitbucket.org/andyfusniakteam/ecom-api-go/utils/nestedset"
	log "github.com/sirupsen/logrus"
)

// PgModel contains the database handle
type PgModel struct {
	db *sql.DB
}

// NewPgModel creates a new PgModel instance
func NewPgModel(db *sql.DB) *PgModel {
	return &PgModel{
		db: db,
	}
}

// CreateCart creates a new shopping cart
func (m *PgModel) CreateCart(ctx context.Context) (*string, error) {
	log.Debug("m.CreateCart() started")

	var cartUUID string
	query := `SELECT UUID_GENERATE_V4() AS uuid`
	err := m.db.QueryRowContext(ctx, query).Scan(&cartUUID)
	if err != nil {
		log.Errorf("db.QueryRow(%s) %+v", query, err)
		return nil, err
	}

	return &cartUUID, nil
}

// AddItemToCart adds a new item with sku, qty and unit price
func (m *PgModel) AddItemToCart(ctx context.Context, cartUUID, tierRef, sku string, qty int) (*model.CartItem, error) {
	item := model.CartItem{}

	// check if the item is alreadyd in the cart
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
		return &item, err
	}

	unitPrice, _ := strconv.ParseFloat(string(unitPriceStr), 64)
	query = `
		INSERT INTO carts (uuid, sku, qty, unit_price)
		VALUES ($1, $2, $3, $4)
		RETURNING id, uuid, sku, qty, unit_price, created, modified
	`

	err = m.db.QueryRowContext(ctx, query, cartUUID, sku, qty, unitPrice).Scan(&item.ID, &item.CartUUID, &item.Sku, &item.Qty, &item.UnitPrice, &item.Created, &item.Modified)
	if err != nil {
		return nil, err
	}

	return &item, nil
}

// GetCartItems gets all items in the cart
func (m *PgModel) GetCartItems(ctx context.Context, cartUUID string) ([]*model.CartItem, error) {
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

	cartItems := make([]*model.CartItem, 0, 20)
	for rows.Next() {
		c := model.CartItem{}
		err = rows.Scan(&c.ID, &c.CartUUID, &c.Sku, &c.Qty, &c.UnitPrice, &c.Created, &c.Modified)
		if err != nil {
			return nil, err
		}

		cartItems = append(cartItems, &c)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return cartItems, nil
}

// UpdateItemByCartUUID updates the qty of a cart item of the given sku.
func (m *PgModel) UpdateItemByCartUUID(ctx context.Context, cartUUID, sku string, qty int) (*model.CartItem, error) {
	query := `
		UPDATE carts
		SET qty = $1, modified = NOW()
		WHERE uuid = $2 AND sku = $3
		RETURNING id, uuid, sku, qty, unit_price, created, modified
	`

	item := model.CartItem{}
	err := m.db.QueryRowContext(ctx, query, qty, cartUUID, sku).Scan(&item.ID, &item.CartUUID, &item.Sku, &item.Qty, &item.UnitPrice, &item.Created, &item.Modified)
	if err != nil {
		return nil, err
	}

	return &item, nil
}

// DeleteCartItem deletes a single cart item.
func (m *PgModel) DeleteCartItem(ctx context.Context, cartUUID, sku string) (count int64, err error) {
	query := `DELETE FROM carts WHERE uuid = $1 AND sku = $2`
	res, err := m.db.ExecContext(ctx, query, cartUUID, sku)
	if err != nil {
		return -1, err
	}
	count, err = res.RowsAffected()
	if err != nil {
		return -1, err
	}

	return count, nil
}

// EmptyCartItems empty the cart of all items. Does not affect coupons.
func (m *PgModel) EmptyCartItems(ctx context.Context, cartUUID string) (err error) {
	query := `DELETE FROM carts WHERE uuid = $1`
	_, err = m.db.ExecContext(ctx, query, cartUUID)
	if err != nil {
		return err
	}

	return nil
}

// CreateCustomer creates a new customer
func (m *PgModel) CreateCustomer(ctx context.Context, UID, email, firstname, lastname string) (*model.Customer, error) {
	c := model.Customer{}
	query := `
		INSERT INTO customers (
			uid, email, firstname, lastname
		) VALUES (
			$1, $2, $3, $4
		)
		RETURNING id, uuid, uid, email, firstname, lastname, created, modified
	`
	err := m.db.QueryRowContext(ctx, query, UID, email, firstname, lastname).Scan(
		&c.ID, &c.CustomerUUID, &c.UID, &c.Email, &c.Firstname, &c.Lastname, &c.Created, &c.Modified)
	if err != nil {
		return nil, err
	}

	return &c, nil
}

// GetCustomers gets the next size customers starting at page page
func (m *PgModel) GetCustomers(ctx context.Context, pq *model.PaginationQuery) (*model.PaginationResultSet, error) {
	q := NewQuery("customers", map[string]bool{
		"id":        true,
		"uuid":      false,
		"uid":       false,
		"email":     true,
		"firstname": true,
		"lastname":  true,
		"created":   true,
		"modified":  true,
	})
	q = q.Select([]string{"id", "uuid", "uid", "email", "firstname", "lastname", "created", "modified"})

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
	pr := model.PaginationResultSet{}
	sql := `
		SELECT COUNT(*) AS count
		FROM %s
	`
	sql = fmt.Sprintf(sql, q.table)
	err := m.db.QueryRowContext(ctx, sql).Scan(&pr.RContext.Total)
	if err != nil {
		return nil, err
	}

	sql = `
		SELECT uuid
		FROM %s
		ORDER BY %s %s, id %s
		FETCH FIRST 1 ROW ONLY
	`
	sql = fmt.Sprintf(sql, q.table, q.orderBy, string(q.orderDir), string(q.orderDir))
	err = m.db.QueryRowContext(ctx, sql).Scan(&pr.RContext.FirstUUID)
	if err != nil {
		return nil, err
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
		return nil, err
	}

	rows, err := m.QueryContextQ(ctx, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	customers := make([]*model.Customer, 0)
	for rows.Next() {
		var c model.Customer

		err = rows.Scan(&c.ID, &c.CustomerUUID, &c.UID, &c.Email, &c.Firstname, &c.Lastname, &c.Created, &c.Modified)
		if err != nil {
			return nil, err
		}

		customers = append(customers, &c)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	pr.RSet = customers
	return &pr, nil
}

// GetCustomerByUUID gets a customer by customer UUID
func (m *PgModel) GetCustomerByUUID(ctx context.Context, customerUUID string) (*model.Customer, error) {
	c := model.Customer{}
	query := `
		SELECT
			id, uuid, uid, email, firstname, lastname, created, modified
		FROM customers
		WHERE uuid = $1
	`
	err := m.db.QueryRowContext(ctx, query, customerUUID).Scan(&c.ID, &c.CustomerUUID, &c.UID, &c.Email, &c.Firstname, &c.Lastname, &c.Created, &c.Modified)
	if err != nil {
		return nil, err
	}

	return &c, nil
}

// CreateAddress creates a new billing or shipping address for a customer
func (m *PgModel) CreateAddress(ctx context.Context, customerID int, typ, contactName, addr1 string, addr2 *string, city string, county *string, postcode, country string) (*model.Address, error) {
	a := model.Address{}
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
		return nil, err
	}

	return &a, nil
}

// GetAddressByUUID gets an address by UUID. Returns a pointer to an Address.
func (m *PgModel) GetAddressByUUID(ctx context.Context, addrUUID string) (*model.Address, error) {
	a := model.Address{}
	query := `
		SELECT
			id, uuid, customer_id, typ, contact_name, addr1, addr2,
			city, county, postcode, country, created, modified
		FROM addresses
		WHERE uuid = $1
	`

	err := m.db.QueryRowContext(ctx, query, addrUUID).Scan(&a.ID, &a.AddrUUID, &a.CustomerID, &a.Typ, &a.ContactName, &a.Addr1, &a.Addr2, &a.City, &a.County, &a.Postcode, &a.Country, &a.Created, &a.Modified)
	if err != nil {
		return nil, err
	}

	return &a, nil
}

// GetAddressOwnerByUUID returns a pointer to a string containing the customer UUID of the owner of this address record. If the address is not found the return value of will be nil.
func (m *PgModel) GetAddressOwnerByUUID(ctx context.Context, addrUUID string) (*string, error) {
	query := `
		SELECT C.uuid
		FROM customers AS C, addresses AS A
		WHERE A.customer_id = C.id AND A.uuid = $1
	`
	var customerUUID string
	err := m.db.QueryRowContext(ctx, query, addrUUID).Scan(&customerUUID)
	if err != nil {
		return nil, err
	}

	return &customerUUID, nil
}

// GetCustomerIDByUUID converts between customer UUID and the underlying primary key
func (m *PgModel) GetCustomerIDByUUID(ctx context.Context, customerUUID string) (int, error) {
	var id int

	query := `SELECT id FROM customers WHERE uuid = $1`
	row := m.db.QueryRowContext(ctx, query, customerUUID)
	err := row.Scan(&id)
	if err != nil {
		return -1, err
	}

	return id, nil
}

// GetAddresses retrieves a slice of pointers to Address for a given customer
func (m *PgModel) GetAddresses(ctx context.Context, customerID int) ([]*model.Address, error) {
	addresses := make([]*model.Address, 0, 8)

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
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var a model.Address

		err = rows.Scan(&a.ID, &a.AddrUUID, &a.CustomerID, &a.Typ, &a.ContactName, &a.Addr1, &a.Addr2, &a.City, &a.County, &a.Postcode, &a.Country, &a.Created, &a.Modified)
		if err != nil {
			return nil, err
		}

		addresses = append(addresses, &a)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return addresses, nil
}

// UpdateAddressByUUID updates an address for a given customer
func (m *PgModel) UpdateAddressByUUID(ctx context.Context, addrUUID string) (*model.Address, error) {
	// TO BE DONE
	//
	//query := `UPDATE addresses SET`
	addr := model.Address{}
	return &addr, nil
}

// DeleteAddressByUUID deletes an address by uuid
func (m *PgModel) DeleteAddressByUUID(ctx context.Context, addrUUID string) error {
	query := `DELETE FROM addresses WHERE uuid = $1`

	_, err := m.db.ExecContext(ctx, query, addrUUID)
	if err != nil {
		return err
	}

	return nil
}

// GetCatalogNestedSet returns a slice of NestedSetNode representing the catalog as a nested set.
func (m *PgModel) GetCatalogNestedSet(ctx context.Context) ([]*nestedset.NestedSetNode, error) {
	query := `
		SELECT id, parent, segment, path, name, lft, rgt, depth, created, modified
		FROM catalog
		ORDER BY lft ASC
	`
	rows, err := m.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	nodes := make([]*nestedset.NestedSetNode, 0, 256)
	for rows.Next() {
		var n nestedset.NestedSetNode
		err = rows.Scan(&n.ID, &n.Parent, &n.Segment, &n.Path, &n.Name, &n.Lft, &n.Rgt, &n.Depth, &n.Created, &n.Modified)
		if err != nil {
			return nil, err
		}
		nodes = append(nodes, &n)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return nodes, nil
}

// GetCatalogProductAssocs returns an Slice of catalogue to product associations
func (m *PgModel) GetCatalogProductAssocs(ctx context.Context) ([]*model.CatalogProductAssoc, error) {
	query := `
		SELECT id, catalog_id, product_id, path, sku, pri
		FROM catalog_products
		ORDER BY pri
	`
	rows, err := m.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	cpa := make([]*model.CatalogProductAssoc, 0, 256)
	for rows.Next() {
		var n model.CatalogProductAssoc
		err = rows.Scan(&n.ID, &n.CatalogID, &n.ProductID, &n.Path, &n.SKU, &n.Pri, &n.Created, &n.Modified)
		if err != nil {
			return nil, err
		}
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return cpa, nil
}

// UpdateCatalogProductAssocs update the catalog product associations
func (m *PgModel) UpdateCatalogProductAssocs(ctx context.Context, cpo []*model.CatalogProductAssoc) error {
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

// CreateImageEntry writes a new image entry to the product_images table
func (m *PgModel) CreateImageEntry(ctx context.Context, c *model.CreateProductImage) (*model.ProductImage, error) {
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
	p := model.ProductImage{}
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
func (m *PgModel) GetImageEntries(ctx context.Context, sku string) ([]*model.ProductImage, error) {
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

	images := make([]*model.ProductImage, 0, 16)
	for rows.Next() {
		p := model.ProductImage{}
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

// ConfirmImageUploaded updates the `up` column to true to indicate the uploaded has taken place.
func (m *PgModel) ConfirmImageUploaded(ctx context.Context, uuid string) (*model.ProductImage, error) {
	query := `
		UPDATE product_images
		SET up = 't', modified = NOW()
		WHERE uuid = $1
		RETURNING id, product_id, uuid, sku, w, h, path, typ, ori, up, pri, size, q,
		gsurl, data, created, modified
	`
	p := model.ProductImage{}
	err := m.db.QueryRowContext(ctx, query, uuid).Scan(&p.ID, &p.ProductID, &p.UUID, &p.SKU, &p.W, &p.H, &p.Path, &p.Typ, &p.Ori, &p.Up, &p.Pri, &p.Size, &p.Q, &p.GSURL, &p.Data, &p.Created, &p.Modified)
	if err != nil {
		return nil, err
	}

	return &p, nil
}

// DeleteImageEntry deletes an image entry row from the product_images table by UUID.
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
