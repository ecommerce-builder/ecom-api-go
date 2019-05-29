package postgres

import (
	"database/sql"
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
