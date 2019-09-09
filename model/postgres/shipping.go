package postgres

import (
	"context"
	"database/sql"
	"time"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

// ErrShippingTarrifCodeExists error for duplicates.
var ErrShippingTarrifCodeExists = errors.New("postgres: shipping tarrif code exists")

// ErrShippingTarrifNotFound error
var ErrShippingTarrifNotFound = errors.New("postgres: shipping tarrif not found")

// ShippingTarrifRow maps to a row in the shipping_tarrif table.
type ShippingTarrifRow struct {
	id           int
	UUID         string
	CountryCode  string
	ShippingCode string
	Name         string
	Price        int
	TaxCode      string
	Created      time.Time
	Modified     time.Time
}

// CreateShippingTarrif attempts to create a new shipping tarrif row.
func (m *PgModel) CreateShippingTarrif(ctx context.Context, countryCode, shippingCode, name string, price int, taxCode string) (*ShippingTarrifRow, error) {
	contextLogger := log.WithContext(ctx)
	contextLogger.Debugf("postgres: CreateShippingTarrif(ctx context.Context, countryCode=%q, shippingCode=%q, name=%q, price=%d, taxCode=%q, ...) started", countryCode, shippingCode, name, price, taxCode)

	// 1. Check if the shipping code exists
	q1 := `SELECT EXISTS(SELECT 1 FROM shipping_tarrif WHERE shipping_code = $1) AS exists`
	var exists bool
	err := m.db.QueryRowContext(ctx, q1, shippingCode).Scan(&exists)
	if err != nil {
		return nil, errors.Wrapf(err, "model: tx.QueryRowContext(ctx, q2=%q, shippingCode=%q)", q1, shippingCode)
	}
	if exists {
		return nil, ErrShippingTarrifCodeExists
	}

	// 2. Insert the new shipping tarrif
	q2 := `
		INSERT INTO shipping_tarrif
		(country_code, shipping_code, name, price, tax_code, created, modified)
		VALUES ($1, $2, $3, $4, $5, NOW(), NOW())
		RETURNING
		  id, uuid, country_code, shipping_code, name, price, tax_code, created, modified
	`
	s := ShippingTarrifRow{}
	row := m.db.QueryRowContext(ctx, q2, countryCode, shippingCode, name, price, taxCode)
	if err := row.Scan(&s.id, &s.UUID, &s.CountryCode, &s.ShippingCode, &s.Name, &s.Price, &s.TaxCode, &s.Created, &s.Modified); err != nil {
		return nil, errors.Wrapf(err, "postgres: query row context q2=%q failed", q2)
	}
	contextLogger.Debugf("postgres: q2 created new shipping tarrif with id=%d, uuid=%s", s.id, s.UUID)

	return &s, nil
}

// GetShippingTarrifByUUID return a single ShippingTarrifRow by uuid.
func (m *PgModel) GetShippingTarrifByUUID(ctx context.Context, shippingTarrifUUID string) (*ShippingTarrifRow, error) {
	q1 := `
		SELECT id, uuid, country_code, shipping_code, name, price, tax_code, created, modified
		FROM shipping_tarrif
		WHERE uuid = $1
	`
	s := ShippingTarrifRow{}
	row := m.db.QueryRowContext(ctx, q1, shippingTarrifUUID)
	if err := row.Scan(&s.id, &s.UUID, &s.CountryCode, &s.ShippingCode, &s.Name, &s.Price, &s.TaxCode, &s.Created, &s.Modified); err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrShippingTarrifNotFound
		}
		return nil, errors.Wrapf(err, "postgres: query scan shippingTarrifUUID=%q q1=%q failed", shippingTarrifUUID, q1)
	}
	return &s, nil
}

// GetShippingTarrifs returns a list of all ShippingTarrifRows from the
// shipping_tarrif table.
func (m *PgModel) GetShippingTarrifs(ctx context.Context) ([]*ShippingTarrifRow, error) {
	q1 := `
		SELECT id, uuid, country_code, shipping_code, name, price, tax_code, created, modified
		FROM shipping_tarrif
	`
	rows, err := m.db.QueryContext(ctx, q1)
	if err != nil {
		return nil, errors.Wrapf(err, "postgres: m.db.QueryContext(ctx) failed")
	}
	defer rows.Close()

	tarrifs := make([]*ShippingTarrifRow, 0, 4)
	for rows.Next() {
		var s ShippingTarrifRow
		if err = rows.Scan(&s.id, &s.UUID, &s.CountryCode, &s.ShippingCode, &s.Name, &s.Price, &s.TaxCode, &s.Created, &s.Modified); err != nil {
			if err == sql.ErrNoRows {
				return nil, ErrShippingTarrifNotFound
			}
			return nil, errors.Wrap(err, "postgres: scan failed")
		}
		tarrifs = append(tarrifs, &s)
	}
	if err = rows.Err(); err != nil {
		return nil, errors.Wrap(err, "postgres: rows.Err()")
	}
	return tarrifs, nil
}

// UpdateShippingTarrif updates a shipping tarrif.
func (m *PgModel) UpdateShippingTarrif(ctx context.Context, shoppingTarrifUUID, countryCode, shippingCode, name string, price int, taxCode string) (*ShippingTarrifRow, error) {
	tx, err := m.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, errors.Wrap(err, "postgres: db.BeginTx")
	}

	q1 := "SELECT id FROM shipping_tarrif WHERE uuid = $1"
	var shippingTarrifID int
	err = tx.QueryRowContext(ctx, q1, shoppingTarrifUUID).Scan(&shippingTarrifID)
	if err != nil {
		if err == sql.ErrNoRows {
			tx.Rollback()
			return nil, ErrShippingTarrifNotFound
		}
		tx.Rollback()
		return nil, errors.Wrapf(err, "postgres: query row context failed for q1=%q", q1)
	}

	// The clause
	//   AND id != $2
	// ensures we are allowed to update our own price list name and description.
	q2 := "SELECT EXISTS(SELECT 1 FROM shipping_tarrif WHERE shipping_code = $1 AND id != $2) AS exists"
	var exists bool
	err = tx.QueryRowContext(ctx, q2, shippingCode, shippingTarrifID).Scan(&exists)
	if err != nil {
		return nil, errors.Wrapf(err, "postgres: tx.QueryRowContext(ctx, q2=%q, shippingCode=%q)", q2, shippingCode)
	}
	if exists {
		return nil, ErrShippingTarrifCodeExists
	}

	q3 := `
		UPDATE shipping_tarrif
		SET country_code = $1, shipping_code = $2, name = $3, price = $4, tax_code = $5, modified = NOW()
		WHERE id = $6
		RETURNING
		  id, uuid, country_code, shipping_code, name, price, tax_code, created, modified
	`
	s := ShippingTarrifRow{}
	row := tx.QueryRowContext(ctx, q3, countryCode, shippingCode, name, price, taxCode, shippingTarrifID)
	if err := row.Scan(&s.id, &s.UUID, &s.CountryCode, &s.ShippingCode, &s.Name, &s.Price, &s.TaxCode, &s.Created, &s.Modified); err != nil {
		tx.Rollback()
		return nil, errors.Wrapf(err, "postgres: query row context q3=%q", q3)
	}

	if err = tx.Commit(); err != nil {
		return nil, errors.Wrap(err, "postgres: tx.Commit failed")
	}
	return &s, nil

}

// DeleteShippingTarrifByUUID deletes a shipping tarrif by uuid.
func (m *PgModel) DeleteShippingTarrifByUUID(ctx context.Context, shippingTarrifUUID string) error {
	tx, err := m.db.BeginTx(ctx, nil)
	if err != nil {
		return errors.Wrap(err, "postgres: db.BeginTx")
	}

	q1 := "SELECT id FROM shipping_tarrif WHERE uuid = $1"
	var shippingTarrifID int
	err = tx.QueryRowContext(ctx, q1, shippingTarrifUUID).Scan(&shippingTarrifID)
	if err != nil {
		if err == sql.ErrNoRows {
			tx.Rollback()
			return ErrShippingTarrifNotFound
		}
		tx.Rollback()

		return errors.Wrapf(err, "postgres: query row context failed for q1=%q", q1)
	}

	q2 := "DELETE FROM shipping_tarrif WHERE id = $1"
	_, err = tx.ExecContext(ctx, q2, shippingTarrifID)
	if err != nil {
		tx.Rollback()
		return errors.Wrapf(err, "postgres: exec context q2=%q", q2)
	}

	if err = tx.Commit(); err != nil {
		return errors.Wrap(err, "postgres: tx.Commit")
	}
	return nil
}
