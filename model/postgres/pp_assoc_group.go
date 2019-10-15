package postgres

import (
	"context"
	"database/sql"
	"time"

	"github.com/pkg/errors"
)

// ErrPPAssocGroupNotFound error.
var ErrPPAssocGroupNotFound = errors.New("postgres: product to product assoc group not found")

// ErrPPAssocGroupExists error
var ErrPPAssocGroupExists = errors.New("postgres: product to product assoc group exists")

// ErrPPAssocGroupContainsAssocs error occurs if attempting to delete a group than
// has existing associations in it.
var ErrPPAssocGroupContainsAssocs = errors.New("postgres: product to product assoc group contains associations")

// PPAssocGroupRow holds a single row from the pp_assoc_group table.
type PPAssocGroupRow struct {
	id       int
	UUID     string
	Code     string
	Name     string
	Created  time.Time
	Modified time.Time
}

// AddPPAssocGroup adds a new row to the pp_assoc_group table.
func (m *PgModel) AddPPAssocGroup(ctx context.Context, code, name string) (*PPAssocGroupRow, error) {
	q1 := `SELECT EXISTS(SELECT 1 FROM pp_assoc_group WHERE code = $1) AS exists`
	var exists bool
	err := m.db.QueryRowContext(ctx, q1, code).Scan(&exists)
	if err != nil {
		return nil, errors.Wrapf(err, "postgres: tx.QueryRowContext(ctx, q1=%q, code=%q)", q1, code)
	}
	if exists {
		return nil, ErrPPAssocGroupExists
	}

	q2 := `
		INSERT INTO pp_assoc_group (code, name)
		VALUES ($1, $2)
		RETURNING id, uuid, code, name, created, modified
	`
	p := PPAssocGroupRow{}
	row := m.db.QueryRowContext(ctx, q2, code, name)
	if err := row.Scan(&p.id, &p.UUID, &p.Code, &p.Name, &p.Created, &p.Modified); err != nil {
		return nil, errors.Wrapf(err, "postgres: query row context q2=%q", q2)
	}

	return &p, nil
}

// GetPPAssocGroup returns a single price list
func (m *PgModel) GetPPAssocGroup(ctx context.Context, ppAssocGroupID string) (*PPAssocGroupRow, error) {
	q1 := `
		SELECT
		  id, uuid, code, name, created, modified
		FROM pp_assoc_group WHERE uuid = $1
	`
	p := PPAssocGroupRow{}
	row := m.db.QueryRowContext(ctx, q1, ppAssocGroupID)
	err := row.Scan(&p.id, &p.UUID, &p.Code, &p.Name, &p.Created, &p.Modified)
	if err == sql.ErrNoRows {
		return nil, ErrPPAssocGroupNotFound
	}
	if err != nil {
		return nil, errors.Wrapf(err, "postgres: scan failed q1=%q ppAssocGroupID=%q", q1, ppAssocGroupID)
	}
	return &p, nil
}

// GetPPAssocGroups returns a list of price lists.
func (m *PgModel) GetPPAssocGroups(ctx context.Context) ([]*PPAssocGroupRow, error) {
	q1 := "SELECT id, uuid, code, name, created, modified FROM pp_assoc_group"
	rows, err := m.db.QueryContext(ctx, q1)
	if err != nil {
		return nil, errors.Wrapf(err, "postgres: m.db.QueryContext(ctx) failed")
	}
	defer rows.Close()

	ppAssocGroups := make([]*PPAssocGroupRow, 0, 2)
	for rows.Next() {
		var p PPAssocGroupRow
		if err = rows.Scan(&p.id, &p.UUID, &p.Code, &p.Name, &p.Created, &p.Modified); err != nil {
			return nil, errors.Wrap(err, "postgres: scan failed")
		}
		ppAssocGroups = append(ppAssocGroups, &p)
	}
	if err := rows.Err(); err != nil {
		return nil, errors.Wrap(err, "postgres: rows.Err()")
	}
	return ppAssocGroups, nil
}

// DeletePPAssocGroup deletes a product to product assoc group by id.
func (m *PgModel) DeletePPAssocGroup(ctx context.Context, ppAssocGroupUUID string) error {
	tx, err := m.db.BeginTx(ctx, nil)
	if err != nil {
		return errors.Wrap(err, "postgres: db.BeginTx failed")
	}

	// 1. Check if the product to product assoc group exists.
	q1 := "SELECT id FROM pp_assoc_group WHERE uuid = $1"
	var ppAssocGroupID int
	err = tx.QueryRowContext(ctx, q1, ppAssocGroupUUID).Scan(&ppAssocGroupID)
	if err == sql.ErrNoRows {
		tx.Rollback()
		return ErrPPAssocGroupNotFound
	}
	if err != nil {
		tx.Rollback()
		return errors.Wrapf(err, "postges: query row context failed for q1=%q", q1)
	}

	// 2. Check if the product to product assoc group already
	// contains associations.
	q2 := "SELECT COUNT(*) AS count FROM pp_assoc WHERE pp_assoc_group_id = $1"
	var count int
	err = tx.QueryRowContext(ctx, q2, ppAssocGroupID).Scan(&count)
	if err != nil {
		tx.Rollback()
		return errors.Wrapf(err, "postgres: scan failed for q2=%q", q2)
	}

	if count > 0 {
		return ErrPPAssocGroupContainsAssocs
	}

	q3 := "DELETE FROM pp_assoc_group WHERE id = $1"
	_, err = tx.ExecContext(ctx, q3, ppAssocGroupID)
	if err != nil {
		tx.Rollback()
		return errors.Wrapf(err, "postgres: exec context query=%q", q3)
	}

	if err := tx.Commit(); err != nil {
		return errors.Wrap(err, "postgres: tx.Commit")
	}
	return nil
}
