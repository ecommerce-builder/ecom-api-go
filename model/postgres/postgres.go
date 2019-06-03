package postgres

import (
	"context"
	"database/sql"

	"github.com/pkg/errors"
)

// PgModel contains the database handle
type PgModel struct {
	db *sql.DB
}

// NewPgModel creates a new PgModel instance
func NewPgModel(db *sql.DB) *PgModel {
	return &PgModel{
		db: db,
	}
}

// GetSchemaVersion returns the underlying schema version string.
func (m *PgModel) GetSchemaVersion(ctx context.Context) (*string, error) {
	query := "SELECT schema_version() AS schema_version"
	var version string
	if err := m.db.QueryRowContext(ctx, query).Scan(&version); err != nil {
		return nil, errors.Wrapf(err, "query scan for schema_version() function failed")
	}
	return &version, nil
}
