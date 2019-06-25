package postgres

import (
	"context"
	"time"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

// A NestedSetNode represents a single node in the nested set.
type NestedSetNode struct {
	id       int
	Segment  string
	Path     string
	Name     string
	Lft      int
	Rgt      int
	Depth    int
	Created  time.Time
	Modified time.Time
}

// BatchCreateNestedSet creates a nested set of nodes representing the
// catalog.
func (m *PgModel) BatchCreateNestedSet(ctx context.Context, ns []*NestedSetNode) error {
	tx, err := m.db.BeginTx(ctx, nil)
	if err != nil {
		return errors.Wrap(err, "db.BeginTx")
	}
	query := "DELETE FROM categories"
	if _, err = tx.ExecContext(ctx, query); err != nil {
		tx.Rollback()
		return errors.Wrapf(err, "model: delete categories query=%q", query)
	}
	query = `
		INSERT INTO categories (
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
func (m *PgModel) GetCatalogByPath(ctx context.Context, path string) (*NestedSetNode, error) {
	query := `
		SELECT id, segment, path, name, lft, rgt, depth, created, modified
		FROM categories
		WHERE path = $1
	`
	var n NestedSetNode
	row := m.db.QueryRowContext(ctx, query, path)
	if err := row.Scan(&n.id, &n.Segment, &n.Path, &n.Name, &n.Lft, &n.Rgt, &n.Depth, &n.Created, &n.Modified); err != nil {
		return nil, errors.Wrapf(err, "service: query row ctx scan query=%q", query)
	}
	return &n, nil
}

// HasCatalog returns true if any rows exist in the categories table.
func (m *PgModel) HasCatalog(ctx context.Context) (bool, error) {
	query := "SELECT COUNT(*) AS count FROM categories"
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
func (m *PgModel) GetCatalogNestedSet(ctx context.Context) ([]*NestedSetNode, error) {
	query := `
		SELECT id, segment, path, name, lft, rgt, depth, created, modified
		FROM categories
		ORDER BY lft ASC
	`
	log.WithContext(ctx).WithFields(log.Fields{
		"query": query,
	}).Debug("SQL query")
	rows, err := m.db.QueryContext(ctx, query)
	if err != nil {
		return nil, errors.Wrapf(err, "query context query=%q", query)
	}
	defer rows.Close()

	nodes := make([]*NestedSetNode, 0, 256)
	for rows.Next() {
		var n NestedSetNode
		if err = rows.Scan(&n.id, &n.Segment, &n.Path, &n.Name, &n.Lft, &n.Rgt, &n.Depth, &n.Created, &n.Modified); err != nil {
			return nil, err
		}
		nodes = append(nodes, &n)
	}

	if err = rows.Err(); err != nil {
		return nil, errors.Wrap(err, "rows err")
	}
	return nodes, nil
}

// DeleteCatalogNestedSet delete all rows in the categories table.
func (m *PgModel) DeleteCatalogNestedSet(ctx context.Context) error {
	query := `DELETE FROM categories`
	_, err := m.db.ExecContext(ctx, query)
	if err != nil {
		return errors.Wrap(err, "service: delete categories")
	}
	return nil
}
