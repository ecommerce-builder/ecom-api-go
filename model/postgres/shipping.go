package postgres

import (
	"context"
	"database/sql"
	"time"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

// ErrShippingTariffCodeExists error for duplicates.
var ErrShippingTariffCodeExists = errors.New("postgres: shipping tariff code exists")

// ErrShippingTariffNotFound error
var ErrShippingTariffNotFound = errors.New("postgres: shipping tariff not found")

// ShippingTariffRow maps to a row in the shipping_tariff table.
type ShippingTariffRow struct {
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

// CreateShippingTariff attempts to create a new shipping tariff row.
func (m *PgModel) CreateShippingTariff(ctx context.Context, countryCode, shippingCode, name string, price int, taxCode string) (*ShippingTariffRow, error) {
	contextLogger := log.WithContext(ctx)
	contextLogger.Debugf("postgres: CreateShippingTariff(ctx context.Context, countryCode=%q, shippingCode=%q, name=%q, price=%d, taxCode=%q, ...) started", countryCode, shippingCode, name, price, taxCode)

	// 1. Check if the shipping code exists
	q1 := `SELECT EXISTS(SELECT 1 FROM shipping_tariff WHERE shipping_code = $1) AS exists`
	var exists bool
	err := m.db.QueryRowContext(ctx, q1, shippingCode).Scan(&exists)
	if err != nil {
		return nil, errors.Wrapf(err, "postgres: tx.QueryRowContext(ctx, q2=%q, shippingCode=%q)", q1, shippingCode)
	}
	if exists {
		return nil, ErrShippingTariffCodeExists
	}

	// 2. Insert the new shipping tariff
	q2 := `
		INSERT INTO shipping_tariff
		(country_code, shipping_code, name, price, tax_code, created, modified)
		VALUES ($1, $2, $3, $4, $5, NOW(), NOW())
		RETURNING
		  id, uuid, country_code, shipping_code, name, price, tax_code, created, modified
	`
	s := ShippingTariffRow{}
	row := m.db.QueryRowContext(ctx, q2, countryCode, shippingCode, name, price, taxCode)
	if err := row.Scan(&s.id, &s.UUID, &s.CountryCode, &s.ShippingCode, &s.Name, &s.Price, &s.TaxCode, &s.Created, &s.Modified); err != nil {
		return nil, errors.Wrapf(err, "postgres: query row context q2=%q failed", q2)
	}
	contextLogger.Debugf("postgres: q2 created new shipping tariff with id=%d, uuid=%s", s.id, s.UUID)

	return &s, nil
}

// GetShippingTariffByUUID return a single ShippingTariffRow by uuid.
func (m *PgModel) GetShippingTariffByUUID(ctx context.Context, shippingTariffUUID string) (*ShippingTariffRow, error) {
	q1 := `
		SELECT
		  id, uuid, country_code, shipping_code, name, price, tax_code,
		  created, modified
		FROM shipping_tariff
		WHERE uuid = $1
	`
	s := ShippingTariffRow{}
	row := m.db.QueryRowContext(ctx, q1, shippingTariffUUID)
	err := row.Scan(&s.id, &s.UUID, &s.CountryCode, &s.ShippingCode, &s.Name, &s.Price,
		&s.TaxCode, &s.Created, &s.Modified)
	if err == sql.ErrNoRows {
		return nil, ErrShippingTariffNotFound
	}
	if err != nil {
		return nil, errors.Wrapf(err, "postgres: query scan shippingTariffUUID=%q q1=%q failed", shippingTariffUUID, q1)
	}
	return &s, nil
}

// GetShippingTariffs returns a list of all ShippingTariffRows from the
// shipping_tariff table.
func (m *PgModel) GetShippingTariffs(ctx context.Context) ([]*ShippingTariffRow, error) {
	q1 := `
		SELECT id, uuid, country_code, shipping_code, name, price, tax_code, created, modified
		FROM shipping_tariff
	`
	rows, err := m.db.QueryContext(ctx, q1)
	if err != nil {
		return nil, errors.Wrapf(err, "postgres: m.db.QueryContext(ctx) failed")
	}
	defer rows.Close()

	tariffs := make([]*ShippingTariffRow, 0, 4)
	for rows.Next() {
		var s ShippingTariffRow
		err = rows.Scan(&s.id, &s.UUID, &s.CountryCode, &s.ShippingCode, &s.Name, &s.Price, &s.TaxCode, &s.Created, &s.Modified)
		if err == sql.ErrNoRows {
			return nil, ErrShippingTariffNotFound
		}
		if err != nil {
			return nil, errors.Wrap(err, "postgres: scan failed")
		}
		tariffs = append(tariffs, &s)
	}
	if err = rows.Err(); err != nil {
		return nil, errors.Wrap(err, "postgres: rows.Err()")
	}
	return tariffs, nil
}

// UpdateShippingTariff updates a shipping tariff.
func (m *PgModel) UpdateShippingTariff(ctx context.Context, shoppingTariffUUID, countryCode, shippingCode, name string, price int, taxCode string) (*ShippingTariffRow, error) {
	tx, err := m.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, errors.Wrap(err, "postgres: db.BeginTx")
	}

	q1 := "SELECT id FROM shipping_tariff WHERE uuid = $1"
	var shippingTariffID int
	err = tx.QueryRowContext(ctx, q1, shoppingTariffUUID).Scan(&shippingTariffID)
	if err == sql.ErrNoRows {
		tx.Rollback()
		return nil, ErrShippingTariffNotFound
	}
	if err != nil {
		tx.Rollback()
		return nil, errors.Wrapf(err, "postgres: query row context failed for q1=%q", q1)
	}

	// The clause
	//   AND id != $2
	// ensures we are allowed to update our own price list name and description.
	q2 := "SELECT EXISTS(SELECT 1 FROM shipping_tariff WHERE shipping_code = $1 AND id != $2) AS exists"
	var exists bool
	err = tx.QueryRowContext(ctx, q2, shippingCode, shippingTariffID).Scan(&exists)
	if err != nil {
		return nil, errors.Wrapf(err,
			"postgres: tx.QueryRowContext(ctx, q2=%q, shippingCode=%q)",
			q2, shippingCode)
	}
	if exists {
		return nil, ErrShippingTariffCodeExists
	}

	q3 := `
		UPDATE shipping_tariff
		SET country_code = $1, shipping_code = $2, name = $3, price = $4, tax_code = $5, modified = NOW()
		WHERE id = $6
		RETURNING
		  id, uuid, country_code, shipping_code, name, price, tax_code, created, modified
	`
	s := ShippingTariffRow{}
	row := tx.QueryRowContext(ctx, q3, countryCode, shippingCode, name, price, taxCode, shippingTariffID)
	if err := row.Scan(&s.id, &s.UUID, &s.CountryCode, &s.ShippingCode, &s.Name, &s.Price, &s.TaxCode, &s.Created, &s.Modified); err != nil {
		tx.Rollback()
		return nil, errors.Wrapf(err, "postgres: query row context q3=%q", q3)
	}

	if err := tx.Commit(); err != nil {
		return nil, errors.Wrap(err, "postgres: tx.Commit failed")
	}
	return &s, nil

}

// DeleteShippingTariffByUUID deletes a shipping tariff by uuid.
func (m *PgModel) DeleteShippingTariffByUUID(ctx context.Context, shippingTariffUUID string) error {
	tx, err := m.db.BeginTx(ctx, nil)
	if err != nil {
		return errors.Wrap(err, "postgres: db.BeginTx")
	}

	q1 := "SELECT id FROM shipping_tariff WHERE uuid = $1"
	var shippingTariffID int
	err = tx.QueryRowContext(ctx, q1, shippingTariffUUID).Scan(&shippingTariffID)
	if err == sql.ErrNoRows {
		tx.Rollback()
		return ErrShippingTariffNotFound
	}
	if err != nil {
		tx.Rollback()
		return errors.Wrapf(err, "postgres: query row context failed for q1=%q", q1)
	}

	q2 := "DELETE FROM shipping_tariff WHERE id = $1"
	_, err = tx.ExecContext(ctx, q2, shippingTariffID)
	if err != nil {
		tx.Rollback()
		return errors.Wrapf(err, "postgres: exec context q2=%q", q2)
	}

	if err := tx.Commit(); err != nil {
		return errors.Wrap(err, "postgres: tx.Commit")
	}
	return nil
}
