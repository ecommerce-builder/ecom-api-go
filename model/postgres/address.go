package postgres

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"time"

	"github.com/lib/pq"
	"github.com/pkg/errors"
)

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

// Value marshals NewAddress to a JSON string.
func (a NewAddress) Value() (driver.Value, error) {
	b, err := json.Marshal(a)
	if err != nil {
		return nil, errors.Wrap(err, "json marshal of NewAddress failed")
	}
	return string(b), nil
}

// Address contains address information for a user
type Address struct {
	id          int
	UUID        string
	usrID       int
	Typ         string
	ContactName string
	Addr1       string
	Addr2       *string
	City        string
	County      *string
	Postcode    string
	Country     string
	Created     time.Time
	Modified    time.Time
}

// CreateAddress creates a new billing or shipping address for a user
func (m *PgModel) CreateAddress(ctx context.Context, userID int, typ, contactName, addr1 string, addr2 *string, city string, county *string, postcode, country string) (*Address, error) {
	a := Address{}
	query := `
		INSERT INTO address (
		  usr_id, typ, contact_name, addr1, addr2, city, county, postcode, country
		) VALUES (
		  $1, $2, $3, $4, $5, $6, $7, $8, $9
		) RETURNING
		  id, uuid, usr_id, typ, contact_name, addr1, addr2, city,
		  county, postcode, country, created, modified
	`
	row := m.db.QueryRowContext(ctx, query, userID, typ, contactName, addr1, addr2, city, county, postcode, country)
	if err := row.Scan(&a.id, &a.UUID, &a.usrID, &a.Typ, &a.ContactName, &a.Addr1,
		&a.Addr2, &a.City, &a.County, &a.Postcode, &a.Country, &a.Created, &a.Modified); err != nil {
		return nil, errors.Wrapf(err, "postgres: query row context scan query=%q", query)
	}
	return &a, nil
}

// GetAddressByUUID gets an address by UUID. Returns a pointer to an Address.
func (m *PgModel) GetAddressByUUID(ctx context.Context, uuid string) (*Address, error) {
	a := Address{}
	query := `
		SELECT
		  id, uuid, usr_id, typ, contact_name, addr1, addr2,
		  city, county, postcode, country, created, modified
		FROM address
		WHERE uuid = $1
	`
	row := m.db.QueryRowContext(ctx, query, uuid)
	if err := row.Scan(&a.id, &a.UUID, &a.usrID, &a.Typ, &a.ContactName, &a.Addr1,
		&a.Addr2, &a.City, &a.County, &a.Postcode, &a.Country, &a.Created, &a.Modified); err != nil {
		if err == sql.ErrNoRows {
			return nil, &ResourceError{
				Op:       "GetAddressByUUID",
				Resource: "address",
				UUID:     uuid,
				Err:      ErrNotExist,
			}
		}
		if pge, ok := err.(*pq.Error); ok {
			switch pge.Code.Name() {
			case "invalid_text_representation":
				return nil, &ResourceError{
					Op:       "GetAddressByUUID",
					Resource: "address",
					UUID:     uuid,
					Err:      ErrInvalidText,
				}
			default:
				return nil, pge
			}
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
func (m *PgModel) GetAddresses(ctx context.Context, userID int) ([]*Address, error) {
	addresses := make([]*Address, 0, 8)
	query := `
		SELECT
		  id, uuid, usr_id, typ, contact_name, addr1,
		  addr2, city, county, postcode, country, created, modified
		FROM address WHERE usr_id = $1
		ORDER BY created DESC
	`
	rows, err := m.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, errors.Wrapf(err, "postgres: db query context query=%q", query)
	}
	defer rows.Close()

	for rows.Next() {
		var a Address
		if err = rows.Scan(&a.id, &a.UUID, &a.usrID, &a.Typ, &a.ContactName, &a.Addr1,
			&a.Addr2, &a.City, &a.County, &a.Postcode, &a.Country, &a.Created, &a.Modified); err != nil {
			return nil, errors.Wrapf(err, "postgres: rows scan query=%q", query)
		}
		addresses = append(addresses, &a)
	}

	if err = rows.Err(); err != nil {
		return nil, errors.Wrap(err, "postgres: rows err")
	}
	return addresses, nil
}

// UpdateAddressByUUID updates an address for a given user
func (m *PgModel) UpdateAddressByUUID(ctx context.Context, UUID string) (*Address, error) {
	// TO BE DONE
	//
	//query := `UPDATE address SET`
	addr := Address{}
	return &addr, nil
}

// DeleteAddressByUUID deletes an address by uuid
func (m *PgModel) DeleteAddressByUUID(ctx context.Context, UUID string) error {
	query := `DELETE FROM address WHERE uuid = $1`
	_, err := m.db.ExecContext(ctx, query, UUID)
	if err != nil {
		return errors.Wrapf(err, "exec context query=%q", query)
	}
	return nil
}
