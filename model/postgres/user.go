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
var ErrUserNotFound = errors.New("postgres: user not found")

// ErrUserInUse is returned when attempting to delete
// a user that has previously placed orders
var ErrUserInUse = errors.New("postgres: user in use")

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
	EndBefore  string
}

// CreateUser creates a new user
func (m *PgModel) CreateUser(ctx context.Context, uid, role, email, firstname, lastname string) (*UsrJoinRow, error) {
	// 1. Look up the default price list
	var priceListID int
	var priceListUUID string
	q1 := "SELECT id, uuid FROM price_list WHERE code = 'default'"
	err := m.db.QueryRowContext(ctx, q1).Scan(&priceListID, &priceListUUID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrDefaultPriceListNotFound
		}
		return nil, errors.Wrapf(err, "postgres: query row context failed for q1=%q", q1)
	}

	q2 := `
		INSERT INTO usr (
		  uid, price_list_id, role, email, firstname, lastname
		) VALUES (
		  $1, $2, $3, $4, $5, $6
		)
		RETURNING
		  id, uuid, uid, price_list_id, role, email, firstname, lastname, created, modified
	`
	u := UsrJoinRow{}
	err = m.db.QueryRowContext(ctx, q2, uid, priceListID, role, email, firstname, lastname).Scan(
		&u.id, &u.UUID, &u.UID, &u.priceListID, &u.Role, &u.Email, &u.Firstname, &u.Lastname, &u.Created, &u.Modified)
	if err != nil {
		return nil, errors.Wrapf(err, "scan failed q2=%q", q2)
	}
	u.PriceListUUID = priceListUUID
	return &u, nil
}

// GetUsers gets the next size user starting at page page
func (m *PgModel) GetUsers(ctx context.Context, pq *PaginationQuery) (*PaginationResultSet, error) {
	q := NewQuery("usr", map[string]bool{
		"id":            true,
		"uuid":          false,
		"uid":           false,
		"price_list_id": false,
		"role":          true,
		"email":         true,
		"firstname":     true,
		"lastname":      true,
		"created":       true,
		"modified":      true,
	})
	q = q.Select([]string{
		"id",
		"uuid",
		"uid",
		"price_list_id",
		"role",
		"email",
		"firstname",
		"lastname",
		"created",
		"modified",
	})

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
	q = q.Limit(pq.Limit)

	if pq.StartAfter != "" {
		q = q.StartAfter(pq.StartAfter)
	}
	if pq.EndBefore != "" {
		q = q.EndBefore(pq.EndBefore)
	}

	// calculate the total count, first and last items in the result set
	pr := PaginationResultSet{}
	sql := "SELECT COUNT(*) AS count FROM %s"
	sql = fmt.Sprintf(sql, q.table)
	err := m.db.QueryRowContext(ctx, sql).Scan(&pr.RContext.Total)
	if err != nil {
		return nil, errors.Wrapf(err, "query row context scan query=%q", sql)
	}

	// book mark either end of the result set
	sql = "SELECT uuid FROM %s ORDER BY %s %s, id %s FETCH FIRST 1 ROW ONLY"
	sql = fmt.Sprintf(sql, q.table, q.orderBy, string(q.orderDir), string(q.orderDir))
	err = m.db.QueryRowContext(ctx, sql).Scan(&pr.RContext.FirstUUID)
	if err != nil {
		return nil, errors.Wrapf(err, "postgres: query row context query=%q", sql)
	}
	sql = "SELECT uuid FROM %s ORDER BY %s %s, id %s FETCH FIRST 1 ROW ONLY"
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
		if err = rows.Scan(&u.id, &u.UUID, &u.UID, &u.priceListID, &u.Role, &u.Email, &u.Firstname, &u.Lastname, &u.Created, &u.Modified); err != nil {
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

// DeleteUserByUUID deletes the usr row with the given uuid.
// Returns the firebase uid of the user.
func (m *PgModel) DeleteUserByUUID(ctx context.Context, usrUUID string) (string, error) {
	tx, err := m.db.BeginTx(ctx, nil)
	if err != nil {
		return "", errors.Wrap(err, "postgres: db.BeginTx")
	}

	// 1. Check the usr exists
	q1 := "SELECT id, uid FROM usr WHERE uuid = $1"
	var usrID int
	var uid string
	err = tx.QueryRowContext(ctx, q1, usrUUID).Scan(&usrID, &uid)
	if err != nil {
		if err == sql.ErrNoRows {
			tx.Rollback()
			return "", ErrUserNotFound
		}

		tx.Rollback()
		return "", errors.Wrapf(err, "postgres: scan failed for q1=%q", q1)
	}

	// 2. Check if the usr is in use (we only really care if they've previously
	// placed an order). Having address and devkeys don't count.
	q2 := `SELECT COUNT(*) AS count FROM "order" WHERE usr_id = $1`
	var count int
	err = tx.QueryRowContext(ctx, q2, usrID).Scan(&count)
	if err != nil {
		tx.Rollback()
		return "", errors.Wrapf(err, "postgres: scan failed for q2=%q", q2)
	}
	if count > 0 {
		tx.Rollback()
		return "", ErrUserInUse
	}

	// 3. Delete any addresses belonging to this user.
	q3 := "DELETE FROM address WHERE usr_id = $1"
	_, err = tx.ExecContext(ctx, q3, usrID)
	if err != nil {
		tx.Rollback()
		return "", errors.Wrapf(err, "postgres: exec context q3=%q", q3)
	}

	// 4. Delete any devkeys belonging to
	q4 := "DELETE FROM usr_devkey WHERE usr_id = $1"
	_, err = tx.ExecContext(ctx, q4, usrID)
	if err != nil {
		tx.Rollback()
		return "", errors.Wrapf(err, "postgres: exec context q4=%q", q4)
	}

	// 5. Delete the usr
	q5 := "DELETE FROM usr WHERE uuid = $1"
	_, err = tx.ExecContext(ctx, q5, usrUUID)
	if err != nil {
		tx.Rollback()
		return "", errors.Wrapf(err, "postgres: exec context q5=%q", q5)
	}

	if err = tx.Commit(); err != nil {
		tx.Rollback()
		return "", errors.Wrap(err, "postgres: tx.Commit")
	}
	return uid, nil
}
