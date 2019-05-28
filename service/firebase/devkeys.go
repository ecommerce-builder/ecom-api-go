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

// CustomerDevKey struct holding the details of a customer Developer Key including its bcrypt hash.
type CustomerDevKey struct {
	UUID         string    `json:"uuid"`
	Key          string    `json:"key"`
	CustomerUUID string    `json:"customer_uuid"`
	Created      time.Time `json:"created"`
	Modified     time.Time `json:"modified"`
}

// GenerateCustomerDevKey creates a new API Key for a customer
func (s *Service) GenerateCustomerDevKey(ctx context.Context, uuid string) (*CustomerDevKey, error) {
	customerID, err := s.model.GetCustomerIDByUUID(ctx, uuid)
	if err != nil {
		return nil, errors.Wrapf(err, "s.model.GetCustomerIDByUUID(ctx, %q)", uuid)
	}
	data := make([]byte, 32)
	_, err = rand.Read(data)
	if err != nil {
		return nil, errors.Wrap(err, "rand.Read(data)")
	}

	ak, err := s.model.CreateCustomerDevKey(ctx, customerID, base58.Encode(data))
	if err != nil {
		return nil, errors.Wrapf(err, "s.model.CreateCustomerDevKey(ctx, customerID=%q, ...)", customerID)
	}

	return &CustomerDevKey{
		Key:      ak.Key,
		Created:  ak.Created,
		Modified: ak.Modified,
	}, nil
}

// GetCustomerDevKey returns a CustomerDevKey for the customer with the given UUID.
func (s *Service) GetCustomerDevKey(ctx context.Context, uuid string) (*CustomerDevKey, error) {
	ak, err := s.model.GetCustomerDevKey(ctx, uuid)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, err
		}
		return nil, err
	}
	return &CustomerDevKey{
		UUID:         ak.UUID,
		Key:          ak.Key,
		CustomerUUID: ak.CustomerUUID,
		Created:      ak.Created,
		Modified:     ak.Modified,
	}, nil
}

// ListCustomersDevKeys gets all API Keys for a customer.
func (s *Service) ListCustomersDevKeys(ctx context.Context, uuid string) ([]*CustomerDevKey, error) {
	customerID, err := s.model.GetCustomerIDByUUID(ctx, uuid)
	if err != nil {
		return nil, err
	}

	rows, err := s.model.GetCustomerDevKeys(ctx, customerID)
	if err != nil {
		return nil, err
	}
	apiKeys := make([]*CustomerDevKey, 0, len(rows))
	for _, row := range rows {
		c := CustomerDevKey{
			UUID:         row.UUID,
			Key:          row.Key,
			CustomerUUID: uuid,
			Created:      row.Created,
			Modified:     row.Modified,
		}
		apiKeys = append(apiKeys, &c)
	}
	return apiKeys, nil
}

// SignInWithDevKey checks the apiKey hash using bcrypt.
func (s *Service) SignInWithDevKey(ctx context.Context, key string) (customToken string, customer *Customer, err error) {
	ak, err := s.model.GetCustomerDevKeyByDevKey(ctx, key)
	if err != nil {
		if err == sql.ErrNoRows {
			// if no key matches create a dummy apiKey struct
			// to ensure the compare hash happens. This mitigates against
			// timing attacks.
			ak = &postgres.CustomerDevKeyFull{
				Key:  "none",
				Hash: "$2a$14$dRgjB9nBHoCs5txdVgN2EeVopE8rfZ7gLJNpLxw9GYq.u53FD00ny", // "nomatch"
			}
		} else {
			return "", nil, errors.Wrap(err, "s.model.GetCustomerDevKeyByDevKey(ctx, key)")
		}
	}

	err = bcrypt.CompareHashAndPassword([]byte(ak.Hash), []byte(ak.Key))
	if err != nil {
		return "", nil, errors.Wrap(err, "bcrypt.CompareHashAndPassword([]byte(ak.Hash), []byte(ak.Key))")
	}

	cust, err := s.model.GetCustomerByUUID(ctx, ak.CustomerUUID)
	if err != nil {
		return "", nil, errors.Wrapf(err, "s.model.GetCustomerByUUID(ctx, customerID=%q)", ak.CustomerUUID)
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
	customer = &Customer{
		UUID:      cust.UUID,
		UID:       cust.UID,
		Role:      cust.Role,
		Email:     cust.Email,
		Firstname: cust.Firstname,
		Lastname:  cust.Lastname,
		Created:   cust.Created,
		Modified:  cust.Modified,
	}
	return token, customer, nil
}
