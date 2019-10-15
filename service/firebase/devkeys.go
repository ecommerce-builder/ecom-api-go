package firebase

import (
	"context"
	"crypto/rand"
	"database/sql"
	"time"

	"bitbucket.org/andyfusniakteam/ecom-api-go/model/postgres"
	"github.com/btcsuite/btcutil/base58"
	"github.com/pkg/errors"
	"golang.org/x/crypto/bcrypt"
)

// ErrDeveloperKeyNotFound error
var ErrDeveloperKeyNotFound = errors.New("service: developer key not found")

// DeveloperKey struct holding the details of a user Developer Key including its bcrypt hash.
type DeveloperKey struct {
	Object   string    `json:"object"`
	ID       string    `json:"id"`
	UserID   string    `json:"user_id"`
	Key      string    `json:"key"`
	Created  time.Time `json:"created"`
	Modified time.Time `json:"modified"`
}

// GenerateUserDevKey creates a new API Key for a user
func (s *Service) GenerateUserDevKey(ctx context.Context, userID string) (*DeveloperKey, error) {
	data := make([]byte, 32)
	_, err := rand.Read(data)
	if err != nil {
		return nil, errors.Wrap(err, "service: rand.Read(data)")
	}

	ak, err := s.model.CreateUserDevKey(ctx, userID, base58.Encode(data))
	if err == postgres.ErrUserNotFound {
		return nil, ErrUserNotFound
	}
	if err != nil {
		return nil, errors.Wrapf(err, "service: s.model.CreateUserDevKey(ctx, userID=%q, ...)", userID)
	}

	return &DeveloperKey{
		Object:   "developer_key",
		ID:       ak.UUID,
		Key:      ak.Key,
		UserID:   userID,
		Created:  ak.Created,
		Modified: ak.Modified,
	}, nil
}

// GetUserDevKey returns a UserDevKey for the user with the given ID.
func (s *Service) GetUserDevKey(ctx context.Context, id string) (*DeveloperKey, error) {
	ak, err := s.model.GetUserDevKey(ctx, id)
	if err == sql.ErrNoRows {
		return nil, err
	}
	if err != nil {
		return nil, err
	}
	return &DeveloperKey{
		Object:   "developer_key",
		ID:       ak.UUID,
		Key:      ak.Key,
		UserID:   ak.UsrUUID,
		Created:  ak.Created,
		Modified: ak.Modified,
	}, nil
}

// ListUsersDevKeys gets all API Keys for a user.
func (s *Service) ListUsersDevKeys(ctx context.Context, userUUID string) ([]*DeveloperKey, error) {
	rows, err := s.model.GetUserDevKeys(ctx, userUUID)
	if err == postgres.ErrUserNotFound {
		return nil, ErrUserNotFound
	}
	if err != nil {
		return nil, err
	}
	apiKeys := make([]*DeveloperKey, 0, len(rows))
	for _, row := range rows {
		c := DeveloperKey{
			Object:   "developer_key",
			ID:       row.UUID,
			Key:      row.Key,
			UserID:   row.UsrUUID,
			Created:  row.Created,
			Modified: row.Modified,
		}
		apiKeys = append(apiKeys, &c)
	}
	return apiKeys, nil
}

// SignInWithDevKey checks the apiKey hash using bcrypt.
func (s *Service) SignInWithDevKey(ctx context.Context, key string) (customToken string, user *User, err error) {
	ak, err := s.model.GetUserDevKeyByDevKey(ctx, key)
	if err == sql.ErrNoRows {
		// if no key matches create a dummy apiKey struct
		// to ensure the compare hash happens. This mitigates against
		// timing attacks.
		ak = &postgres.UsrDevKeyJoinRow{
			Key:  "none",
			Hash: "$2a$14$dRgjB9nBHoCs5txdVgN2EeVopE8rfZ7gLJNpLxw9GYq.u53FD00ny", // "nomatch"
		}
	}
	if err != nil {
		return "", nil, errors.Wrap(err, "service: s.model.GetUserDevKeyByDevKey(ctx, key)")
	}

	err = bcrypt.CompareHashAndPassword([]byte(ak.Hash), []byte(ak.Key))
	if err != nil {
		return "", nil, errors.Wrap(err, "bcrypt.CompareHashAndPassword([]byte(ak.Hash), []byte(ak.Key))")
	}

	cust, err := s.model.GetUserByUUID(ctx, ak.UsrUUID)
	if err != nil {
		return "", nil, errors.Wrapf(err, "s.model.GetUserByUUID(ctx, userID=%q)", ak.UsrUUID)
	}

	authClient, err := s.fbApp.Auth(ctx)
	if err != nil {
		return "", nil, errors.Wrap(err, "s.fbApp.Auth(ctx)")
	}

	userRecord, err := authClient.GetUserByEmail(ctx, cust.Email)
	if err != nil {
		return "", nil, errors.Wrapf(err, "authClient.GetUserByEmail(ctx, email=%q)", cust.Email)
	}

	token, err := authClient.CustomToken(ctx, userRecord.UID)
	if err != nil {
		return "", nil, errors.Wrapf(err, "authClient.CustomToken(ctx, uid=%q)", userRecord.UID)
	}
	user = &User{
		ID:        cust.UUID,
		UID:       cust.UID,
		Role:      cust.Role,
		Email:     cust.Email,
		Firstname: cust.Firstname,
		Lastname:  cust.Lastname,
		Created:   cust.Created,
		Modified:  cust.Modified,
	}
	return token, user, nil
}

// DeleteDeveloperKey deletes the developer key with the given id.
func (s *Service) DeleteDeveloperKey(ctx context.Context, developerKeyID string) error {
	err := s.model.DeleteUsrDevKey(ctx, developerKeyID)
	if err == postgres.ErrDeveloperKeyNotFound {
		return ErrDeveloperKeyNotFound
	}
	if err != nil {
		return errors.Wrapf(err, "service: s.model.DeleteUsrDevKey(ctx, usrDevKeyUUID=%q) failed", developerKeyID)
	}
	return nil
}
