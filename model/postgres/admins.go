package postgres

import (
	"context"

	"github.com/pkg/errors"
)

// GetAdmin returns a Customer of role admin for the given UUID.
func (m *PgModel) GetAdmin(ctx context.Context, uuid string) (*Customer, error) {
	query := `
		SELECT
			id, uuid, uid, role, email, firstname, lastname, created, modified
		FROM customers
		WHERE uuid = $1 AND role='admin'
	`
	c := Customer{}
	row := m.db.QueryRowContext(ctx, query, uuid)
	if err := row.Scan(&c.id, &c.UUID, &c.UID, &c.Role, &c.Email, &c.Firstname, &c.Lastname, &c.Created, &c.Modified); err != nil {
		return nil, errors.Wrapf(err, "query row context scan query=%q Customer=%v", query, c)
	}
	return &c, nil
}

// GetAllAdmins returns a slice of Customers who are all of role admin
func (m *PgModel) GetAllAdmins(ctx context.Context) ([]*Customer, error) {
	query := `
		SELECT id, uuid, uid, role, email, firstname, lastname, created, modified
		FROM customers
		WHERE role = 'admin'
		ORDER by created DESC
	`
	rows, err := m.db.QueryContext(ctx, query)
	if err != nil {
		return nil, errors.Wrapf(err, "m.db.QueryContext(ctx) query=%q", query)
	}
	defer rows.Close()

	admins := make([]*Customer, 0, 8)
	for rows.Next() {
		var c Customer
		if err := rows.Scan(&c.id, &c.UUID, &c.UID, &c.Role, &c.Email, &c.Firstname, &c.Lastname, &c.Created, &c.Modified); err != nil {
			return nil, errors.Wrap(err, "scan failed")
		}
		admins = append(admins, &c)
	}
	if err = rows.Err(); err != nil {
		return nil, errors.Wrap(err, "rows.Err()")
	}
	return admins, nil
}

// IsAdmin returns true is the given customer UUID has a role of admin.
func (m *PgModel) IsAdmin(ctx context.Context, uuid string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM customers WHERE uuid=$1 AND role='admin') AS exists`
	var exists bool
	err := m.db.QueryRowContext(ctx, query, uuid).Scan(&exists)
	if err != nil {
		return false, errors.Wrapf(err, "db.QueryRow(ctx, %s)", query)
	}
	return exists, nil
}

// DeleteAdminByUUID deletes the administrator from the customers table
// with the given UUID.
func (m *PgModel) DeleteAdminByUUID(ctx context.Context, uuid string) error {
	query := `DELETE FROM customers WHERE uuid = $1 AND role = 'admin'`
	_, err := m.db.ExecContext(ctx, query, uuid)
	if err != nil {
		return errors.Wrapf(err, "exec context query=%q", query)
	}
	return nil
}
