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

// UserDevKey struct holding the details of a user Developer Key including its bcrypt hash.
type UserDevKey struct {
	ID       string    `json:"id"`
	Key      string    `json:"key"`
	UserID   string    `json:"user_id,omitempty"`
	Created  time.Time `json:"created"`
	Modified time.Time `json:"modified"`
}

// GenerateUserDevKey creates a new API Key for a user
func (s *Service) GenerateUserDevKey(ctx context.Context, userID string) (*UserDevKey, error) {
	uid, err := s.model.GetUserIDByUUID(ctx, userID)
	if err != nil {
		if err == postgres.ErrUserNotFound {
			return nil, ErrUserNotFound
		}
		return nil, errors.Wrapf(err, "service: s.model.GetUserIDByUUID(ctx, userID=%q)", userID)
	}
	data := make([]byte, 32)
	_, err = rand.Read(data)
	if err != nil {
		return nil, errors.Wrap(err, "service: rand.Read(data)")
	}

	ak, err := s.model.CreateUserDevKey(ctx, uid, base58.Encode(data))
	if err != nil {
		return nil, errors.Wrapf(err, "service: s.model.CreateUserDevKey(ctx, userID=%q, ...)", userID)
	}

	return &UserDevKey{
		ID:  ak.UUID,
		Key: ak.Key,
		// UserID: userID,
		Created:  ak.Created,
		Modified: ak.Modified,
	}, nil
}

// GetUserDevKey returns a UserDevKey for the user with the given ID.
func (s *Service) GetUserDevKey(ctx context.Context, id string) (*UserDevKey, error) {
	ak, err := s.model.GetUserDevKey(ctx, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, err
		}
		return nil, err
	}
	return &UserDevKey{
		ID:       ak.UUID,
		Key:      ak.Key,
		UserID:   ak.UserUUID,
		Created:  ak.Created,
		Modified: ak.Modified,
	}, nil
}

// ListUsersDevKeys gets all API Keys for a user.
func (s *Service) ListUsersDevKeys(ctx context.Context, userID string) ([]*UserDevKey, error) {
	uid, err := s.model.GetUserIDByUUID(ctx, userID)
	if err != nil {
		if err == postgres.ErrUserNotFound {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	rows, err := s.model.GetUserDevKeys(ctx, uid)
	if err != nil {
		return nil, err
	}
	apiKeys := make([]*UserDevKey, 0, len(rows))
	for _, row := range rows {
		c := UserDevKey{
			ID:       row.UUID,
			Key:      row.Key,
			UserID:   userID,
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
	if err != nil {
		if err == sql.ErrNoRows {
			// if no key matches create a dummy apiKey struct
			// to ensure the compare hash happens. This mitigates against
			// timing attacks.
			ak = &postgres.UsrDevKeyFull{
				Key:  "none",
				Hash: "$2a$14$dRgjB9nBHoCs5txdVgN2EeVopE8rfZ7gLJNpLxw9GYq.u53FD00ny", // "nomatch"
			}
		} else {
			return "", nil, errors.Wrap(err, "s.model.GetUserDevKeyByDevKey(ctx, key)")
		}
	}

	err = bcrypt.CompareHashAndPassword([]byte(ak.Hash), []byte(ak.Key))
	if err != nil {
		return "", nil, errors.Wrap(err, "bcrypt.CompareHashAndPassword([]byte(ak.Hash), []byte(ak.Key))")
	}

	cust, err := s.model.GetUserByUUID(ctx, ak.UserUUID)
	if err != nil {
		return "", nil, errors.Wrapf(err, "s.model.GetUserByUUID(ctx, userID=%q)", ak.UserUUID)
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
