package postgres

import (
	"context"

	"bitbucket.org/andyfusniakteam/ecom-api-go/utils/nestedset"
	"github.com/pkg/errors"
)

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
