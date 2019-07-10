package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/pkg/errors"
)

// CustomerRow holds details of a single row from the customers table.
type CustomerRow struct {
	id        int
	UUID      string
	UID       string
	Role      string
	Email     string
	Firstname string
	Lastname  string
	Created   time.Time
	Modified  time.Time
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

// PaginationQuery for querying the database with paging.
type PaginationQuery struct {
	OrderBy    string
	OrderDir   string
	Limit      int
	StartAfter string
}

// CreateCustomer creates a new customer
func (m *PgModel) CreateCustomer(ctx context.Context, uid, role, email, firstname, lastname string) (*CustomerRow, error) {
	query := `
		INSERT INTO customers (
			uid, role, email, firstname, lastname
		) VALUES (
			$1, $2, $3, $4, $5
		)
		RETURNING id, uuid, uid, role, email, firstname, lastname, created, modified
	`
	c := CustomerRow{}
	err := m.db.QueryRowContext(ctx, query, uid, role, email, firstname, lastname).Scan(
		&c.id, &c.UUID, &c.UID, &c.Role, &c.Email, &c.Firstname, &c.Lastname, &c.Created, &c.Modified)
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

	customers := make([]*CustomerRow, 0)
	for rows.Next() {
		var c CustomerRow
		if err = rows.Scan(&c.id, &c.UUID, &c.UID, &c.Role, &c.Email, &c.Firstname, &c.Lastname, &c.Created, &c.Modified); err != nil {
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
func (m *PgModel) GetCustomerByUUID(ctx context.Context, customerUUID string) (*CustomerRow, error) {
	query := `
		SELECT
			id, uuid, uid, role, email, firstname, lastname, created, modified
		FROM customers
		WHERE uuid = $1
	`
	c := CustomerRow{}
	row := m.db.QueryRowContext(ctx, query, customerUUID)
	if err := row.Scan(&c.id, &c.UUID, &c.UID, &c.Role, &c.Email, &c.Firstname, &c.Lastname, &c.Created, &c.Modified); err != nil {
		return nil, errors.Wrapf(err, "query row context scan query=%q Customer=%v", query, c)
	}
	return &c, nil
}

// GetCustomerByID gets a customer by customer ID
func (m *PgModel) GetCustomerByID(ctx context.Context, customerID int) (*CustomerRow, error) {
	query := `
		SELECT
			id, uuid, uid, role, email, firstname, lastname, created, modified
		FROM customers
		WHERE id = $1
	`
	c := CustomerRow{}
	row := m.db.QueryRowContext(ctx, query, customerID)
	if err := row.Scan(&c.id, &c.UUID, &c.UID, &c.Role, &c.Email, &c.Firstname, &c.Lastname, &c.Created, &c.Modified); err != nil {
		return nil, errors.Wrapf(err, "query row context scan query=%q Customer=%v", query, c)
	}
	return &c, nil
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
