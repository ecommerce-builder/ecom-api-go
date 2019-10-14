package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/pkg/errors"
)

// ErrAddressNotFound error
var ErrAddressNotFound = errors.New("postgres: address not found")

// NewAddress contains details of a new address to add to the database.
type NewAddress struct {
	ContactName string
	Addr1       string
	Addr2       string
	City        string
	County      string
	Postcode    string
	Country     string
}

// AddressJoinRow contains address information for a user
type AddressJoinRow struct {
	id          int
	UUID        string
	usrID       int
	UsrUUID     string
	Typ         string
	ContactName string
	Addr1       string
	Addr2       *string
	City        string
	County      *string
	Postcode    string
	CountryCode string
	Created     time.Time
	Modified    time.Time
}

// CreateAddress creates a new billing or shipping address for a user
func (m *PgModel) CreateAddress(ctx context.Context, usrUUID, typ, contactName, addr1 string, addr2 *string, city string, county *string, postcode, countryCode string) (*AddressJoinRow, error) {
	// 1. Check the user exists
	q1 := "SELECT id FROM usr WHERE uuid = $1"
	var usrID int
	err := m.db.QueryRowContext(ctx, q1, usrUUID).Scan(&usrID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrUserNotFound
		}
		return nil, errors.Wrapf(err, "postgres: query row context failed q1=%q", q1)
	}

	// 2. Insert the new address
	var a AddressJoinRow
	q2 := `
		INSERT INTO address
		  (usr_id, typ, contact_name, addr1, addr2, city, county, postcode, country_code)
		VALUES
		  ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING
		  id, uuid, usr_id, typ, contact_name, addr1, addr2, city,
		  county, postcode, country_code, created, modified
	`
	row := m.db.QueryRowContext(ctx, q2, usrID, typ, contactName, addr1, addr2, city, county, postcode, countryCode)
	if err := row.Scan(&a.id, &a.UUID, &a.usrID, &a.Typ, &a.ContactName, &a.Addr1,
		&a.Addr2, &a.City, &a.County, &a.Postcode, &a.CountryCode, &a.Created, &a.Modified); err != nil {
		return nil, errors.Wrapf(err, "postgres: query row context scan q2=%q", q2)
	}
	a.UsrUUID = usrUUID
	return &a, nil
}

// GetAddressByUUID gets an address by UUID. Returns a pointer to an Address.
func (m *PgModel) GetAddressByUUID(ctx context.Context, addressUUID string) (*AddressJoinRow, error) {

	// 1. Check the address exists
	q1 := "SELECT id FROM address WHERE uuid = $1"
	var addressID int
	err := m.db.QueryRowContext(ctx, q1, addressUUID).Scan(&addressID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrAddressNotFound
		}
		return nil, errors.Wrapf(err, "postgres: query row context failed q1=%q", q1)
	}

	a := AddressJoinRow{}
	query := `
		SELECT
		  a.id, a.uuid, usr_id, u.uuid as usr_uuid, typ, contact_name, addr1,
		  addr2, city, county, postcode, country_code, a.created, a.modified
		FROM address AS a
		INNER JOIN usr AS u
		  ON u.id = a.usr_id
		WHERE a.id = $1
	`
	row := m.db.QueryRowContext(ctx, query, addressID)
	if err := row.Scan(&a.id, &a.UUID, &a.usrID, &a.UsrUUID, &a.Typ, &a.ContactName, &a.Addr1,
		&a.Addr2, &a.City, &a.County, &a.Postcode, &a.CountryCode, &a.Created, &a.Modified); err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrAddressNotFound
		}
		return nil, errors.Wrapf(err, "query row context scan query=%q", query)
	}
	return &a, nil
}

// GetAddressOwnerByUUID returns a pointer to a string containing the
// user UUID of the owner of this address record. If the address is not
// found the return value of will be nil.
func (m *PgModel) GetAddressOwnerByUUID(ctx context.Context, uuid string) (*string, error) {
	query := `
		SELECT C.uuid
		FROM usr AS C, address AS A
		WHERE A.usr_id = C.id AND A.uuid = $1
	`
	var userUUID string
	err := m.db.QueryRowContext(ctx, query, uuid).Scan(&userUUID)
	if err != nil {
		return nil, errors.Wrapf(err, "query row context query=%q", query)
	}
	return &userUUID, nil
}

// GetAddresses retrieves a slice of pointers to Address for a given
// user.
func (m *PgModel) GetAddresses(ctx context.Context, usrUUID string) ([]*AddressJoinRow, error) {
	// 1. Check the usr exists
	q1 := "SELECT id FROM usr WHERE uuid = $1"
	var usrID int
	err := m.db.QueryRowContext(ctx, q1, usrUUID).Scan(&usrID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrUserNotFound
		}
		return nil, errors.Wrapf(err, "postgres: query row context failed q1=%q", q1)
	}

	// 2. Get all the address rows for this usr.
	addresses := make([]*AddressJoinRow, 0, 8)
	q2 := `
		SELECT
		  id, uuid, usr_id, typ, contact_name, addr1,
		  addr2, city, county, postcode, country_code, created, modified
		FROM address AS a
		WHERE usr_id = $1
		ORDER BY created DESC
	`
	rows, err := m.db.QueryContext(ctx, q2, usrID)
	if err != nil {
		return nil, errors.Wrapf(err, "postgres: db query context q2=%q", q2)
	}
	defer rows.Close()

	for rows.Next() {
		var a AddressJoinRow
		if err = rows.Scan(&a.id, &a.UUID, &a.usrID, &a.Typ, &a.ContactName, &a.Addr1,
			&a.Addr2, &a.City, &a.County, &a.Postcode, &a.CountryCode, &a.Created, &a.Modified); err != nil {
			return nil, errors.Wrapf(err, "postgres: rows scan q2=%q", q2)
		}
		a.UsrUUID = usrUUID
		addresses = append(addresses, &a)
	}

	if err = rows.Err(); err != nil {
		return nil, errors.Wrap(err, "postgres: rows err")
	}
	return addresses, nil
}

// PartialUpdateAddress updates on or more columns in a single address row
func (m *PgModel) PartialUpdateAddress(ctx context.Context, addressUUID string, typ, contactName, addr1, addr2, city, county, postcode, countryCode *string) (*AddressJoinRow, error) {
	q1 := "SELECT id FROM address WHERE uuid = $1"
	var addressID int
	err := m.db.QueryRowContext(ctx, q1, addressUUID).Scan(&addressID)
	if err == sql.ErrNoRows {
		return nil, ErrAddressNotFound
	}
	if err != nil {
		return nil, errors.Wrapf(err, "postgres: query row context failed q1=%q", q1)
	}

	var set []string
	var queryArgs []interface{}
	argCounter := 1
	if typ != nil {
		set = append(set, fmt.Sprintf("typ = $%d", argCounter))
		argCounter++
		queryArgs = append(queryArgs, *typ)
	}
	if contactName != nil {
		set = append(set, fmt.Sprintf("contact_name = $%d", argCounter))
		argCounter++
		queryArgs = append(queryArgs, *contactName)
	}
	if addr1 != nil {
		set = append(set, fmt.Sprintf("addr1 = $%d", argCounter))
		argCounter++
		queryArgs = append(queryArgs, *addr1)
	}
	if addr2 != nil {
		set = append(set, fmt.Sprintf("addr2 = $%d", argCounter))
		argCounter++
		queryArgs = append(queryArgs, *addr2)
	}
	if city != nil {
		set = append(set, fmt.Sprintf("city = $%d", argCounter))
		argCounter++
		queryArgs = append(queryArgs, *city)
	}
	if county != nil {
		set = append(set, fmt.Sprintf("county = $%d", argCounter))
		argCounter++
		queryArgs = append(queryArgs, *county)
	}
	if postcode != nil {
		set = append(set, fmt.Sprintf("postcode = $%d", argCounter))
		argCounter++
		queryArgs = append(queryArgs, *postcode)
	}
	if countryCode != nil {
		set = append(set, fmt.Sprintf("country_code = $%d", argCounter))
		argCounter++
		queryArgs = append(queryArgs, *countryCode)
	}

	queryArgs = append(queryArgs, addressID)
	setQuery := strings.Join(set, ", ")
	q2 := `
		UPDATE address
		SET
		  %SET_QUERY%, modified = NOW()
		WHERE id = %ARG_COUNTER%
		RETURNING
		  id, uuid, usr_id,
		  (SELECT uuid FROM usr WHERE id = usr_id),
		  typ, contact_name, addr1, addr2, city,
		  county, postcode, country_code, created, modified
	`
	q2 = strings.Replace(q2, "%SET_QUERY%", setQuery, 1)
	q2 = strings.Replace(q2, "%ARG_COUNTER%", fmt.Sprintf("$%d", argCounter), 1)
	a := AddressJoinRow{}
	row := m.db.QueryRowContext(ctx, q2, queryArgs...)
	if err := row.Scan(&a.id, &a.UUID, &a.usrID, &a.UsrUUID, &a.Typ, &a.ContactName, &a.Addr1,
		&a.Addr2, &a.City, &a.County, &a.Postcode, &a.CountryCode, &a.Created, &a.Modified); err != nil {
		return nil, errors.Wrapf(err, "postgres: query row context q2=%q", q2)
	}

	return &a, nil
}

// DeleteAddressByUUID deletes an address by uuid
func (m *PgModel) DeleteAddressByUUID(ctx context.Context, addressUUID string) error {
	q1 := "SELECT id FROM address WHERE uuid = $1"
	var addressID int
	err := m.db.QueryRowContext(ctx, q1, addressUUID).Scan(&addressID)
	if err != nil {
		if err == sql.ErrNoRows {
			return ErrAddressNotFound
		}
		return errors.Wrapf(err, "postgres: query row context failed q1=%q", q1)
	}

	q2 := "DELETE FROM address WHERE id = $1"
	_, err = m.db.ExecContext(ctx, q2, addressID)
	if err != nil {
		return errors.Wrapf(err, "postgres: exec context q2=%q", q2)
	}
	return nil
}
