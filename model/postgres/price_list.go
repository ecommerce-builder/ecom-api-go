package postgres

import (
	"context"
	"database/sql"
	"time"

	"github.com/pkg/errors"
)

// ErrDefaultPriceListMissing error
var ErrDefaultPriceListMissing = errors.New("postgres: pricing list missing")

// ErrPriceListNotFound error
var ErrPriceListNotFound = errors.New("postgres: price list not found")

// ErrPriceListCodeTaken error
var ErrPriceListCodeTaken = errors.New("postgres: price list code already taken")

// ErrPriceListInUse error
var ErrPriceListInUse = errors.New("postgres: price list has associated prices")

// PriceListRow represents a row in the price_list table.
type PriceListRow struct {
	id           int
	UUID         string
	Code         string
	CurrencyCode string
	Strategy     string
	IncTax       bool
	Name         string
	Description  string
	Created      time.Time
	Modified     time.Time
}

// CreatePriceList writes a new price list row to the price_list table
// returning a PriceListRow.
func (m *PgModel) CreatePriceList(ctx context.Context, code, currencyCode, strategy string, incTax bool, name, description string) (*PriceListRow, error) {
	tx, err := m.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, errors.Wrap(err, "db.BeginTx")
	}

	q1 := `SELECT EXISTS(SELECT 1 FROM price_list WHERE code = $1) AS exists`
	var exists bool
	err = tx.QueryRowContext(ctx, q1, code).Scan(&exists)
	if err != nil {
		return nil, errors.Wrapf(err, "postgres: tx.QueryRowContext(ctx, q1=%q, code=%q)", q1, code)
	}

	if exists {
		return nil, ErrPriceListCodeTaken
	}

	q2 := `
		INSERT INTO price_list (code, currency_code, strategy, inc_tax, name, description)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING
		  id, uuid, code, currency_code, strategy, inc_tax, name, description, created, modified
	`
	p := PriceListRow{}
	row := tx.QueryRowContext(ctx, q2, code, currencyCode, strategy, incTax, name, description)
	if err := row.Scan(&p.id, &p.UUID, &p.Code, &p.CurrencyCode, &p.Strategy, &p.IncTax, &p.Name, &p.Description, &p.Created, &p.Modified); err != nil {
		tx.Rollback()
		return nil, errors.Wrapf(err, "query row context q2=%q", q2)
	}

	if err = tx.Commit(); err != nil {
		return nil, errors.Wrap(err, "tx.Commit failed")
	}
	return &p, nil
}

// GetPriceList returns a single price list
func (m *PgModel) GetPriceList(ctx context.Context, priceListUUID string) (*PriceListRow, error) {
	q1 := `
		SELECT
		  id, uuid, code, currency_code, strategy, inc_tax, name, description, created, modified
		FROM price_list WHERE uuid = $1
	`
	p := PriceListRow{}
	row := m.db.QueryRowContext(ctx, q1, priceListUUID)
	if err := row.Scan(&p.id, &p.UUID, &p.Code, &p.CurrencyCode, &p.Strategy, &p.IncTax, &p.Name, &p.Description, &p.Created, &p.Modified); err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrPriceListNotFound
		}
		return nil, errors.Wrapf(err, "query row context scan query=%q priceListUUID=%q", q1, priceListUUID)
	}
	return &p, nil
}

// GetPriceLists returns a list of price lists.
func (m *PgModel) GetPriceLists(ctx context.Context) ([]*PriceListRow, error) {
	q1 := `
		SELECT
		  id, uuid, code, currency_code, strategy, inc_tax, name,
		  description, created, modified
		FROM price_list
	`
	rows, err := m.db.QueryContext(ctx, q1)
	if err != nil {
		return nil, errors.Wrapf(err, "m.db.QueryContext(ctx) failed")
	}
	defer rows.Close()

	tiers := make([]*PriceListRow, 0, 4)
	for rows.Next() {
		var p PriceListRow
		if err = rows.Scan(&p.id, &p.UUID, &p.Code, &p.CurrencyCode, &p.Strategy, &p.IncTax, &p.Name, &p.Description, &p.Created, &p.Modified); err != nil {
			if err == sql.ErrNoRows {
				return nil, ErrPriceListNotFound
			}
			return nil, errors.Wrap(err, "scan failed")
		}
		tiers = append(tiers, &p)
	}
	if err = rows.Err(); err != nil {
		return nil, errors.Wrap(err, "rows.Err()")
	}
	return tiers, nil
}

// UpdatePriceList updates a price list by price list uuid
func (m *PgModel) UpdatePriceList(ctx context.Context, priceListUUID, code, currencyCode, strategy string, incTax bool, name, description string) (*PriceListRow, error) {
	tx, err := m.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, errors.Wrap(err, "model: db.BeginTx")
	}

	q1 := "SELECT id FROM price_list WHERE uuid = $1"
	var priceListID int
	err = tx.QueryRowContext(ctx, q1, priceListUUID).Scan(&priceListID)
	if err != nil {
		if err == sql.ErrNoRows {
			tx.Rollback()
			return nil, ErrPriceListNotFound
		}
		tx.Rollback()
		return nil, errors.Wrapf(err, "model: query row context failed for q1=%q", q1)
	}

	// The clause
	//   AND id != $2
	// ensures we are allowed to update our own price list name and description.
	q2 := `SELECT EXISTS(SELECT 1 FROM price_list WHERE code = $1 AND id != $2) AS exists`
	var exists bool
	err = tx.QueryRowContext(ctx, q2, code, priceListID).Scan(&exists)
	if err != nil {
		return nil, errors.Wrapf(err, "model: tx.QueryRowContext(ctx, q2=%q, code=%q)", q2, code)
	}
	if exists {
		return nil, ErrPriceListCodeTaken
	}

	q3 := `
		UPDATE price_list
		SET
		  code = $1, currency_code = $2, strategy = $3, inc_tax = $4,
		  name = $5, description = $6, modified = NOW()
		WHERE id = $7
		RETURNING
		  id, uuid, code, currency_code, strategy, inc_tax, name, description, created, modified
	`
	p := PriceListRow{}
	row := tx.QueryRowContext(ctx, q3, code, currencyCode, strategy, incTax, name, description, priceListID)
	if err := row.Scan(&p.id, &p.UUID, &p.Code, &p.CurrencyCode, &p.Strategy, &p.IncTax, &p.Name, &p.Description, &p.Created, &p.Modified); err != nil {
		tx.Rollback()
		return nil, errors.Wrapf(err, "model: query row context q3=%q", q3)
	}

	if err = tx.Commit(); err != nil {
		return nil, errors.Wrap(err, "model: tx.Commit failed")
	}
	return &p, nil
}

// DeletePriceList deletes a price list by price list uuid.
func (m *PgModel) DeletePriceList(ctx context.Context, priceListUUID string) error {
	tx, err := m.db.BeginTx(ctx, nil)
	if err != nil {
		return errors.Wrap(err, "model: db.BeginTx")
	}

	q1 := "SELECT id FROM price_list WHERE uuid = $1"
	var priceListID int
	err = tx.QueryRowContext(ctx, q1, priceListUUID).Scan(&priceListID)
	if err != nil {
		if err == sql.ErrNoRows {
			tx.Rollback()
			return ErrPriceListNotFound
		}
		tx.Rollback()
		return errors.Wrapf(err, "model: query row context failed for q1=%q", q1)
	}

	q2 := "SELECT COUNT(*) AS count FROM price WHERE price_list_id = $1"
	var count int
	err = tx.QueryRowContext(ctx, q2, priceListID).Scan(&count)
	if err != nil {
		tx.Rollback()
		return errors.Wrapf(err, "model: query row context failed for q2=%q", q2)
	}

	if count > 0 {
		return ErrPriceListInUse
	}

	q3 := "DELETE FROM price_list WHERE id = $1"
	_, err = tx.ExecContext(ctx, q3, priceListID)
	if err != nil {
		tx.Rollback()
		return errors.Wrapf(err, "model: exec context query=%q", q3)
	}

	if err = tx.Commit(); err != nil {
		return errors.Wrap(err, "model: tx.Commit")
	}
	return nil
}
