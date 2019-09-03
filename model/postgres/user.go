package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/pkg/errors"
)

// ErrUserNotFound is returned when a query for the user
// could not be found in the database.
var ErrUserNotFound = errors.New("model: user not found")

// UsrRow holds details of a single row from the usr table.
type UsrRow struct {
	id          int
	UUID        string
	UID         string
	priceListID int
	Role        string
	Email       string
	Firstname   string
	Lastname    string
	Created     time.Time
	Modified    time.Time
}

// UsrJoinRow joines with the price_list table to use its uuid.
type UsrJoinRow struct {
	id            int
	UUID          string
	UID           string
	priceListID   int
	PriceListUUID string
	Role          string
	Email         string
	Firstname     string
	Lastname      string
	Created       time.Time
	Modified      time.Time
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

// CreateUser creates a new user
func (m *PgModel) CreateUser(ctx context.Context, uid, role, email, firstname, lastname string) (*UsrRow, error) {
	query := `
		INSERT INTO usr (
		  uid, role, email, firstname, lastname
		) VALUES (
		  $1, $2, $3, $4, $5
		)
		RETURNING
		  id, uuid, uid, role, email, firstname, lastname, created, modified
	`
	u := UsrRow{}
	err := m.db.QueryRowContext(ctx, query, uid, role, email, firstname, lastname).Scan(
		&u.id, &u.UUID, &u.UID, &u.Role, &u.Email, &u.Firstname, &u.Lastname, &u.Created, &u.Modified)
	if err != nil {
		return nil, errors.Wrapf(err, "query row context User=%v", u)
	}
	return &u, nil
}

// GetUsers gets the next size user starting at page page
func (m *PgModel) GetUsers(ctx context.Context, pq *PaginationQuery) (*PaginationResultSet, error) {
	q := NewQuery("usr", map[string]bool{
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
		return nil, errors.Wrapf(err, "postgres: query row context query=%q", sql)
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
		return nil, errors.Wrapf(err, "postgres: query row context query=%q", sql)
	}

	rows, err := m.QueryContextQ(ctx, q)
	if err != nil {
		return nil, errors.Wrapf(err, "postgres: model query context q=%v", q)
	}
	defer rows.Close()

	usrs := make([]*UsrRow, 0)
	for rows.Next() {
		var u UsrRow
		if err = rows.Scan(&u.id, &u.UUID, &u.UID, &u.Role, &u.Email, &u.Firstname, &u.Lastname, &u.Created, &u.Modified); err != nil {
			return nil, errors.Wrapf(err, "postgres: rows scan User=%v", u)
		}
		usrs = append(usrs, &u)
	}

	if err = rows.Err(); err != nil {
		return nil, errors.Wrap(err, "postgres: rows err")
	}
	pr.RSet = usrs
	return &pr, nil
}

// GetUserByUUID gets a user by user UUID
func (m *PgModel) GetUserByUUID(ctx context.Context, userUUID string) (*UsrJoinRow, error) {
	query := `
		SELECT
		  c.id, c.uuid, uid, price_list_id, l.uuid as price_list_uuid, role, email,
		  firstname, lastname, c.created, c.modified
		FROM usr AS c
		INNER JOIN price_list AS l
		  ON c.price_list_id = l.id
		WHERE c.uuid = $1
	`
	c := UsrJoinRow{}
	row := m.db.QueryRowContext(ctx, query, userUUID)
	if err := row.Scan(&c.id, &c.UUID, &c.UID, &c.priceListID, &c.PriceListUUID, &c.Role, &c.Email,
		&c.Firstname, &c.Lastname, &c.Created, &c.Modified); err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrUserNotFound
		}
		return nil, errors.Wrapf(err, "query row context scan query=%q User=%v", query, c)
	}
	return &c, nil
}

// GetUserByID gets a user by user ID
func (m *PgModel) GetUserByID(ctx context.Context, userID int) (*UsrRow, error) {
	query := `
		SELECT
		  id, uuid, uid, role, email, firstname, lastname, created, modified
		FROM usr
		WHERE id = $1
	`
	u := UsrRow{}
	row := m.db.QueryRowContext(ctx, query, userID)
	if err := row.Scan(&u.id, &u.UUID, &u.UID, &u.Role, &u.Email, &u.Firstname, &u.Lastname, &u.Created, &u.Modified); err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrUserNotFound
		}
		return nil, errors.Wrapf(err, "query row context scan query=%q User=%v", query, u)
	}
	return &u, nil
}

// GetUserIDByUUID converts between user UUID and the underlying
// primary key.
func (m *PgModel) GetUserIDByUUID(ctx context.Context, userUUID string) (int, error) {
	var id int
	query := `SELECT id FROM usr WHERE uuid = $1`
	row := m.db.QueryRowContext(ctx, query, userUUID)
	err := row.Scan(&id)
	if err != nil {
		if err == sql.ErrNoRows {
			return -1, ErrUserNotFound
		}
		return -1, errors.Wrapf(err, "postgres: query row context query=%q", query)
	}
	return id, nil
}
