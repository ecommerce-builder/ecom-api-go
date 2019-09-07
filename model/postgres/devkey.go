package postgres

import (
	"context"
	"database/sql"
	"time"

	"github.com/pkg/errors"
	"golang.org/x/crypto/bcrypt"
)

// UsrDevKey user developer keys.
type UsrDevKey struct {
	id       int
	UUID     string
	Key      string
	Hash     string
	usrID    int
	Created  time.Time
	Modified time.Time
}

// UsrDevKeyFull contains fields from the usr_devkey joined to usr.
type UsrDevKeyFull struct {
	id       int
	UUID     string
	Key      string
	Hash     string
	UserUUID string
	Created  time.Time
	Modified time.Time
}

// ErrDeveloperKeyNotFound error
var ErrDeveloperKeyNotFound = errors.New("postgres: developer key not found")

// CreateUserDevKey generates a user developer key using bcrypt.
func (m *PgModel) CreateUserDevKey(ctx context.Context, userID int, key string) (*UsrDevKey, error) {
	query := `
		INSERT INTO usr_devkey (
		  key, hash, usr_id, created, modified
		) VALUES (
		  $1, $2, $3, NOW(), NOW()
		) RETURNING
		  id, uuid, key, hash, usr_id, created, modified
	`
	row := UsrDevKey{}
	hash, err := bcrypt.GenerateFromPassword([]byte(key), 14)
	err = m.db.QueryRowContext(ctx, query, key, string(hash), userID).Scan(&row.id, &row.UUID, &row.Key, &row.Hash, &row.usrID, &row.Created, &row.Modified)
	if err != nil {
		return nil, errors.Wrapf(err, "postgres: query row context scan query=%q", query)
	}
	return &row, nil
}

// GetUserDevKeys returns a slice of UsrDevKeys by user primary key.
func (m *PgModel) GetUserDevKeys(ctx context.Context, usrID int) ([]*UsrDevKey, error) {
	query := `
		SELECT
		  id, uuid, key, hash, usr_id, created, modified
		FROM usr_devkey
		WHERE usr_id = $1
	`
	rows, err := m.db.QueryContext(ctx, query, usrID)
	if err != nil {
		return nil, errors.Wrapf(err, "postgres: m.db.QueryContext(ctx, %q, %d)", query, usrID)
	}
	defer rows.Close()

	apiKeys := make([]*UsrDevKey, 0, 8)
	for rows.Next() {
		var row UsrDevKey
		err = rows.Scan(&row.id, &row.UUID, &row.Key, &row.Hash, &row.usrID, &row.Created, &row.Modified)
		if err != nil {
			return nil, errors.Wrap(err, "postgres: scan failed")
		}
		apiKeys = append(apiKeys, &row)
	}
	if err = rows.Err(); err != nil {
		return nil, errors.Wrap(err, "postgres: rows.Err()")
	}
	return apiKeys, nil
}

// GetUserDevKey returns a single UserDevKey by UUID.
func (m *PgModel) GetUserDevKey(ctx context.Context, uuid string) (*UsrDevKeyFull, error) {
	query := `
		SELECT d.id, d.uuid, key, hash, usr_id, u.uuid, d.created, d.modified
		FROM usr_devkey AS d
		INNER JOIN usr AS u ON d.usr_id = u.id
		WHERE A.uuid = $1
	`
	row := UsrDevKeyFull{}
	err := m.db.QueryRowContext(ctx, query, uuid).Scan(&row.id, &row.UUID, &row.Key, &row.Hash, &row.UserUUID, &row.Created, &row.Modified)
	if err != nil {
		return nil, errors.Wrapf(err, "postgres: m.db.QueryRowContext(ctx, %q, %q).Scan(...)", query, uuid)
	}
	return &row, nil
}

// GetUserDevKeyByDevKey retrieves a given Developer Key record.
func (m *PgModel) GetUserDevKeyByDevKey(ctx context.Context, key string) (*UsrDevKeyFull, error) {
	query := `
		SELECT
		  A.id as id, A.uuid as uuid, key, hash,
		  C.uuid as usr_uuid, A.created as created, A.modified as modified
		FROM usr_devkey AS A
		INNER JOIN usr AS C ON A.usr_id = C.id
		WHERE key = $1
	`
	row := UsrDevKeyFull{}
	err := m.db.QueryRowContext(ctx, query, key).Scan(&row.id, &row.UUID, &row.Key, &row.Hash, &row.UserUUID, &row.Created, &row.Modified)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, err
		}
		return nil, errors.Wrapf(err, "postgres: m.db.QueryRowContext(ctx, %q, %q).Scan(...)", query, key)
	}
	return &row, nil
}

// DeleteUsrDevKey deletes the developer key with the given uuid.
func (m *PgModel) DeleteUsrDevKey(ctx context.Context, usrDevKeyUUID string) error {
	q1 := "SELECT id FROM usr_devkey WHERE uuid = $1"
	var usrDevKeyID int
	err := m.db.QueryRowContext(ctx, q1, usrDevKeyUUID).Scan(&usrDevKeyID)
	if err != nil {
		if err == sql.ErrNoRows {
			return ErrDeveloperKeyNotFound
		}
		return errors.Wrapf(err, "postgres: query row context failed for q1=%q", q1)
	}

	q2 := "DELETE FROM usr_devkey WHERE id = $1"
	_, err = m.db.ExecContext(ctx, q2, usrDevKeyID)
	if err != nil {
		return errors.Wrapf(err, "postgres: exec context failed q2=%q", q2)
	}

	return nil
}
