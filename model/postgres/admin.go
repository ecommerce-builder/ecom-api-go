package postgres

import (
	"context"

	"github.com/pkg/errors"
)

// GetAdmin returns a user of role admin for the given UUID.
func (m *PgModel) GetAdmin(ctx context.Context, uuid string) (*UsrRow, error) {
	query := `
		SELECT
		  id, uuid, uid, role, email, firstname, lastname, created, modified
		FROM usr WHERE uuid = $1 AND role = 'admin'
	`
	u := UsrRow{}
	row := m.db.QueryRowContext(ctx, query, uuid)
	if err := row.Scan(&u.id, &u.UUID, &u.UID, &u.Role, &u.Email, &u.Firstname, &u.Lastname, &u.Created, &u.Modified); err != nil {
		return nil, errors.Wrapf(err, "postgres: query row context scan query=%q User=%v", query, u)
	}
	return &u, nil
}

// GetAllAdmins returns a slice of users who are all of role admin
func (m *PgModel) GetAllAdmins(ctx context.Context) ([]*UsrRow, error) {
	query := `
		SELECT
		  id, uuid, uid, role, email, firstname, lastname, created, modified
		FROM usr WHERE role = 'admin'
		ORDER by created DESC
	`
	rows, err := m.db.QueryContext(ctx, query)
	if err != nil {
		return nil, errors.Wrapf(err, "postgres: m.db.QueryContext(ctx) query=%q", query)
	}
	defer rows.Close()

	admins := make([]*UsrRow, 0, 8)
	for rows.Next() {
		var c UsrRow
		if err := rows.Scan(&c.id, &c.UUID, &c.UID, &c.Role, &c.Email, &c.Firstname, &c.Lastname, &c.Created, &c.Modified); err != nil {
			return nil, errors.Wrap(err, "postgres: scan failed")
		}
		admins = append(admins, &c)
	}
	if err = rows.Err(); err != nil {
		return nil, errors.Wrap(err, "postgres: rows.Err()")
	}
	return admins, nil
}

// IsAdmin returns true is the given user UUID has a role of admin.
func (m *PgModel) IsAdmin(ctx context.Context, uuid string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM usr WHERE uuid = $1 AND role = 'admin') AS exists`
	var exists bool
	err := m.db.QueryRowContext(ctx, query, uuid).Scan(&exists)
	if err != nil {
		return false, errors.Wrapf(err, "postgres: db.QueryRow(ctx, %s)", query)
	}
	return exists, nil
}

// DeleteAdminByUUID deletes the administrator from the usr table
// with the given UUID.
func (m *PgModel) DeleteAdminByUUID(ctx context.Context, uuid string) error {
	query := `DELETE FROM usr WHERE uuid = $1 AND role = 'admin'`
	_, err := m.db.ExecContext(ctx, query, uuid)
	if err != nil {
		return errors.Wrapf(err, "postgres: exec context query=%q", query)
	}
	return nil
}
