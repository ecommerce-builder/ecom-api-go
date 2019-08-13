package postgres

import (
	"context"
	"database/sql"
	"time"

	"github.com/pkg/errors"
	"golang.org/x/crypto/bcrypt"
)

// CustomerDevKey customer developer keys.
type CustomerDevKey struct {
	id         int
	UUID       string
	Key        string
	Hash       string
	customerID int
	Created    time.Time
	Modified   time.Time
}

// CustomerDevKeyFull contains fields from the customer_devkey joined to customer.
type CustomerDevKeyFull struct {
	id           int
	UUID         string
	Key          string
	Hash         string
	CustomerUUID string
	Created      time.Time
	Modified     time.Time
}

// CreateCustomerDevKey generates a customer developer key using bcrypt.
func (m *PgModel) CreateCustomerDevKey(ctx context.Context, customerID int, key string) (*CustomerDevKey, error) {
	query := `
		INSERT INTO customer_devkey (
			key, hash, customer_id, created, modified
		) VALUES (
			$1, $2, $3, NOW(), NOW()
		) RETURNING
			id, uuid, key, hash, customer_id, created, modified
	`
	row := CustomerDevKey{}
	hash, err := bcrypt.GenerateFromPassword([]byte(key), 14)
	err = m.db.QueryRowContext(ctx, query, key, string(hash), customerID).Scan(&row.id, &row.UUID, &row.Key, &row.Hash, &row.customerID, &row.Created, &row.Modified)
	if err != nil {
		return nil, errors.Wrapf(err, "query row context scan query=%q", query)
	}
	return &row, nil
}

// GetCustomerDevKeys returns a slice of CustomerDevKeys by customer primary key.
func (m *PgModel) GetCustomerDevKeys(ctx context.Context, customerID int) ([]*CustomerDevKey, error) {
	query := `
		SELECT
		  id, uuid, key, hash, customer_id, created, modified
		FROM customer_devkey
		WHERE customer_id = $1
	`
	rows, err := m.db.QueryContext(ctx, query, customerID)
	if err != nil {
		return nil, errors.Wrapf(err, "m.db.QueryContext(ctx, %q, %d)", query, customerID)
	}
	defer rows.Close()

	apiKeys := make([]*CustomerDevKey, 0, 8)
	for rows.Next() {
		var row CustomerDevKey
		err = rows.Scan(&row.id, &row.UUID, &row.Key, &row.Hash, &row.customerID, &row.Created, &row.Modified)
		if err != nil {
			return nil, errors.Wrap(err, "Scan failed")
		}
		apiKeys = append(apiKeys, &row)
	}
	if err = rows.Err(); err != nil {
		return nil, errors.Wrap(err, "rows.Err()")
	}
	return apiKeys, nil
}

// GetCustomerDevKey returns a single CustomerDevKey by UUID.
func (m *PgModel) GetCustomerDevKey(ctx context.Context, uuid string) (*CustomerDevKeyFull, error) {
	query := `
		SELECT A.id, A.uuid, key, hash, customer_id, C.uuid, A.created, A.modified
		FROM customer_devkey AS A
		INNER JOIN customer AS C ON A.customer_id = C.id
		WHERE
			  A.uuid = $1
	`
	row := CustomerDevKeyFull{}
	err := m.db.QueryRowContext(ctx, query, uuid).Scan(&row.id, &row.UUID, &row.Key, &row.Hash, &row.CustomerUUID, &row.Created, &row.Modified)
	if err != nil {
		return nil, errors.Wrapf(err, " m.db.QueryRowContext(ctx, %q, %q).Scan(...)", query, uuid)
	}
	return &row, nil
}

// GetCustomerDevKeyByDevKey retrieves a given Developer Key record.
func (m *PgModel) GetCustomerDevKeyByDevKey(ctx context.Context, key string) (*CustomerDevKeyFull, error) {
	query := `
		SELECT
		  A.id as id, A.uuid as uuid, key, hash,
		  C.uuid as customer_uuid, A.created as created, A.modified as modified
		FROM customer_devkey AS A
		INNER JOIN customer AS C ON A.customer_id = C.id
		WHERE key = $1
	`
	row := CustomerDevKeyFull{}
	err := m.db.QueryRowContext(ctx, query, key).Scan(&row.id, &row.UUID, &row.Key, &row.Hash, &row.CustomerUUID, &row.Created, &row.Modified)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, err
		}
		return nil, errors.Wrapf(err, "m.db.QueryRowContext(ctx, %q, %q).Scan(...)", query, key)
	}
	return &row, nil
}
