package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
)

// OrderDirection is a named type for ASC or DESC
type OrderDirection string

// Query defines a query for pagination on a given table
type Query struct {
	table      string
	sfields    map[string]bool
	sel        []string
	orderBy    string
	orderDir   OrderDirection
	limit      int
	startAfter string
	err        error
}

func (o OrderDirection) toggle() OrderDirection {
	if o == "ASC" {
		return "DESC"
	}
	return "ASC"
}

// NewQuery builds a new query for paginating a table
func NewQuery(table string, sf map[string]bool) *Query {
	return &Query{
		table:      table,
		sfields:    sf,
		sel:        make([]string, 0, 16),
		orderBy:    "id",
		orderDir:   OrderDirection("ASC"),
		limit:      0,
		startAfter: "",
		err:        nil,
	}
}

// Select indicates which columns to select
func (q *Query) Select(s []string) *Query {
	return &Query{
		table:      q.table,
		sfields:    q.sfields,
		sel:        s,
		orderBy:    q.orderBy,
		orderDir:   q.orderDir,
		limit:      q.limit,
		startAfter: q.startAfter,
		err:        nil,
	}
}

// OrderBy indicates which column to sort on
func (q *Query) OrderBy(o string) *Query {
	return &Query{
		table:      q.table,
		sfields:    q.sfields,
		sel:        q.sel,
		orderBy:    o,
		orderDir:   q.orderDir,
		limit:      q.limit,
		startAfter: q.startAfter,
		err:        q.err,
	}
}

// OrderDir determines the ordering of the result set.
func (q *Query) OrderDir(d OrderDirection) *Query {
	return &Query{
		table:      q.table,
		sfields:    q.sfields,
		sel:        q.sel,
		orderBy:    q.orderBy,
		orderDir:   OrderDirection(strings.ToUpper(string(d))),
		limit:      q.limit,
		startAfter: q.startAfter,
		err:        q.err,
	}
}

// Limit restricts the number of results in the result set to a maximum of n.
func (q *Query) Limit(n int) *Query {
	return &Query{
		table:      q.table,
		sfields:    q.sfields,
		sel:        q.sel,
		orderBy:    q.orderBy,
		orderDir:   q.orderDir,
		limit:      n,
		startAfter: q.startAfter,
		err:        q.err,
	}
}

// StartAfter indicates the UUID of the row to return the next page of results.
func (q *Query) StartAfter(s string) *Query {
	return &Query{
		table:      q.table,
		sfields:    q.sfields,
		sel:        q.sel,
		orderBy:    q.orderBy,
		orderDir:   q.orderDir,
		limit:      q.limit,
		startAfter: s,
		err:        q.err,
	}
}

// QueryContextQ builds an SQL statement based on the given Query q and
// returns make a call to sql.QueryContext returning a *sql.Rows.
func (m *PgModel) QueryContextQ(ctx context.Context, q *Query) (*sql.Rows, error) {
	for _, v := range q.sel {
		_, ok := q.sfields[v]
		if !ok {
			return nil, fmt.Errorf("query: Select contains non searchable field %q", v)
		}
	}

	// if q.startVal == "" {
	// 	return nil, errors.New("query: StartAt/StartAfter must be called with at least one value")
	// }

	if q.orderDir != "ASC" && q.orderDir != "DESC" {
		return nil, errors.New(`query: OrderDirection must be called with either "ASC" or "DESC"`)
	}

	if len(q.sel) == 0 {
		q.sel = append(q.sel, "*")
	}

	// Assume startAfter only
	var comparitor string
	if q.startAfter != "" {
		if q.orderDir == "ASC" {
			comparitor = ">"
		} else {
			comparitor = "<"
		}
	}

	sql := `
		SELECT %s FROM %s
		WHERE (%s, id) %s (
			SELECT %s, id FROM customers
		 	WHERE uuid = $1
		)
		ORDER BY %s %s, id %s
		FETCH FIRST %d ROWS ONLY
	`
	sql = fmt.Sprintf(sql,
		strings.Join(q.sel, ", "), q.table,
		q.orderBy, comparitor,
		q.orderBy,
		q.orderBy, q.orderDir, q.orderDir,
		q.limit)
	rows, err := m.db.QueryContext(ctx, sql, q.startAfter)
	if err != nil {
		return nil, fmt.Errorf("QueryContext with %q: %v", sql, err)
	}
	return rows, err
}
