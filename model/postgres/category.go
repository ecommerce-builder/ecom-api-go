package postgres

import (
	"context"
	"time"

	"github.com/pkg/errors"
)

// ErrCategoryNotFound error
var ErrCategoryNotFound = errors.New("postgres: category not found")

// ErrCategoryNotLeaf error
var ErrCategoryNotLeaf = errors.New("postgres: category not a leaf")

// ErrLeafCategoryNotFound error
var ErrLeafCategoryNotFound = errors.New("postgres: category not found")

// ErrCategoriesInUse error
var ErrCategoriesInUse = errors.New("postgres: categories in use")

// A CategoryRow represents a single row from the category table.
type CategoryRow struct {
	id       int
	UUID     string
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
func (m *PgModel) BatchCreateNestedSet(ctx context.Context, ns []*CategoryRow) error {
	tx, err := m.db.BeginTx(ctx, nil)
	if err != nil {
		return errors.Wrap(err, "postgres: db.BeginTx")
	}

	// 1. Check if any categories are in use with the promo rules
	q1 := `
		SELECT COUNT(*) AS count
		FROM promo_rule
		WHERE category_id IS NOT NULL
	`
	var count int
	err = tx.QueryRowContext(ctx, q1).Scan(&count)
	if err != nil {
		return errors.Wrapf(err, "postgres: tx.QueryRowContext(ctx, q1=%q) failed", q1)
	}
	if count > 0 {
		return ErrCategoriesInUse
	}

	// 2. Delete the all existing categories
	q2 := "DELETE FROM category"
	if _, err = tx.ExecContext(ctx, q2); err != nil {
		tx.Rollback()
		return errors.Wrapf(err, "postgres: delete category q2=%q", q2)
	}

	// 3. Prepare a statment to insert a new category row
	q3 := `
		INSERT INTO category (
		  segment, path, name, lft, rgt, depth, created, modified
		) VALUES (
		  $1, $2, $3, $4, $5, $6, NOW(), NOW()
		)
	`
	stmt3, err := tx.PrepareContext(ctx, q3)
	if err != nil {
		tx.Rollback()
		return errors.Wrapf(err, "postgres: tx prepare for q3=%q", q3)
	}
	defer stmt3.Close()

	for _, n := range ns {
		if _, err := stmt3.ExecContext(ctx, n.Segment, n.Path, n.Name, n.Lft, n.Rgt, n.Depth); err != nil {
			tx.Rollback() // return an error too, we may want to wrap them
			return errors.Wrapf(err,
				"postgres: stmt3 exec segment=%q path=%q name=%q lft=%d rgt=%d depth=%d",
				n.Segment, n.Path, n.Name, n.Lft, n.Rgt, n.Depth)
		}
	}
	if err := tx.Commit(); err != nil {
		return errors.Wrap(err, "postgres: tx.Commit")
	}
	return nil
}

// GetCategoryByPath retrieves a single set element by the given path.
func (m *PgModel) GetCategoryByPath(ctx context.Context, path string) (*CategoryRow, error) {
	query := `
		SELECT
		  id, uuid, segment, path, name, lft, rgt, depth, created, modified
		FROM category
		WHERE path = $1
	`
	var n CategoryRow
	row := m.db.QueryRowContext(ctx, query, path)
	if err := row.Scan(&n.id, &n.UUID, &n.Segment, &n.Path, &n.Name, &n.Lft, &n.Rgt, &n.Depth, &n.Created, &n.Modified); err != nil {
		return nil, errors.Wrapf(err, "service: query row ctx scan query=%q", query)
	}
	return &n, nil
}

// HasCatalog returns true if any rows exist in the category table.
func (m *PgModel) HasCatalog(ctx context.Context) (bool, error) {
	query := "SELECT COUNT(*) AS count FROM category"
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

// GetCategories returns a slice of CategoryRow representing the catalog as a nested set.
func (m *PgModel) GetCategories(ctx context.Context) ([]*CategoryRow, error) {
	query := `
		SELECT
		  id, uuid, segment, path, name, lft, rgt, depth, created, modified
		FROM category
		ORDER BY lft ASC
	`
	rows, err := m.db.QueryContext(ctx, query)
	if err != nil {
		return nil, errors.Wrapf(err, "query context query=%q", query)
	}
	defer rows.Close()

	nodes := make([]*CategoryRow, 0, 256)
	for rows.Next() {
		var n CategoryRow
		if err = rows.Scan(&n.id, &n.UUID, &n.Segment, &n.Path, &n.Name, &n.Lft, &n.Rgt, &n.Depth, &n.Created, &n.Modified); err != nil {
			return nil, err
		}
		nodes = append(nodes, &n)
	}

	if err = rows.Err(); err != nil {
		return nil, errors.Wrap(err, "rows err")
	}
	return nodes, nil
}

// DeleteCategories deletes all rows in the category table.
func (m *PgModel) DeleteCategories(ctx context.Context) error {
	query := "DELETE FROM category"
	_, err := m.db.ExecContext(ctx, query)
	if err != nil {
		return errors.Wrapf(err, "postgres: m.db.ExecContext(ctx, query=%q) failed", query)
	}
	return nil
}
