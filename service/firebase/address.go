package firebase

import (
	"context"
	"time"

	"bitbucket.org/andyfusniakteam/ecom-api-go/model/postgres"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

// ErrAddressNotFound error
var ErrAddressNotFound = errors.New("service: address not found")

// Address contains address information for a user
type Address struct {
	Object      string    `json:"object"`
	ID          string    `json:"id"`
	UserID      string    `json:"user_id"`
	Typ         string    `json:"type"`
	ContactName string    `json:"contact_name"`
	Addr1       string    `json:"addr1"`
	Addr2       *string   `json:"addr2,omitempty"`
	City        string    `json:"city"`
	County      *string   `json:"county,omitempty"`
	Postcode    string    `json:"postcode"`
	CountryCode string    `json:"country_code"`
	Created     time.Time `json:"created"`
	Modified    time.Time `json:"modified"`
}

// CreateAddress creates a new address for a user
func (s *Service) CreateAddress(ctx context.Context, userID, typ, contactName, addr1 string, addr2 *string, city string, countyCode *string, postcode string, country string) (*Address, error) {
	contextLogger := log.WithContext(ctx)
	contextLogger.Infof("service: CreateAddress(ctx, userID=%q, contactName=%q, ...) started",
		userID, contactName)

	a, err := s.model.CreateAddress(ctx, userID, typ, contactName,
		addr1, addr2, city, countyCode, postcode, country)
	if err == postgres.ErrUserNotFound {
		return nil, ErrUserNotFound
	}
	if err != nil {
		return nil, errors.Wrapf(err, "service: s.model.CreateAddress(ctx, userUUID=%q, typ=%q, contactName=%q, addr1=%q, addr2=%v, city=%q, county=%v, postcode=%q, countryCode=%q) failed", userID, typ, contactName, addr1, addr2, city, countyCode, postcode, country)
	}

	addr := Address{
		Object:      "address",
		ID:          a.UUID,
		UserID:      a.UsrUUID,
		Typ:         a.Typ,
		ContactName: a.ContactName,
		Addr1:       a.Addr1,
		Addr2:       a.Addr2,
		City:        a.City,
		County:      a.County,
		Postcode:    a.Postcode,
		CountryCode: a.CountryCode,
		Created:     a.Created,
		Modified:    a.Modified,
	}

	if err := s.PublishTopicEvent(ctx, EventAddressCreated, &addr); err != nil {
		return nil, errors.Wrapf(err,
			"service: s.PublishTopicEvent(ctx, event=%q, data=%v) failed",
			EventAddressCreated, addr)
	}
	contextLogger.WithFields(log.Fields{
		"Object":      "address",
		"ID":          a.UUID,
		"UserID":      a.UsrUUID,
		"Typ":         a.Typ,
		"ContactName": a.ContactName,
		"Addr1":       a.Addr1,
		"Addr2":       a.Addr2,
		"City":        a.City,
		"County":      a.County,
		"Postcode":    a.Postcode,
		"CountryCode": a.CountryCode,
		"Created":     a.Created,
		"Modified":    a.Modified,
	}).Infof("service: EventAddressCreated published")
	return &addr, nil
}

// GetAddress return a single Address.
func (s *Service) GetAddress(ctx context.Context, addressID string) (*Address, error) {
	a, err := s.model.GetAddressByUUID(ctx, addressID)
	if err == postgres.ErrAddressNotFound {
		return nil, ErrAddressNotFound
	}
	if err != nil {
		return nil, errors.Wrapf(err, "s.model.GetAddressByUUID(ctx, addressID=%q) failed", addressID)
	}
	addr := Address{
		Object:      "address",
		ID:          a.UUID,
		UserID:      a.UsrUUID,
		Typ:         a.Typ,
		ContactName: a.ContactName,
		Addr1:       a.Addr1,
		Addr2:       a.Addr2,
		City:        a.City,
		County:      a.County,
		Postcode:    a.Postcode,
		CountryCode: a.CountryCode,
		Created:     a.Created,
		Modified:    a.Modified,
	}
	return &addr, nil
}

// GetAddressOwner returns the user that owns the address with the given ID.
func (s *Service) GetAddressOwner(ctx context.Context, id string) (*string, error) {
	userUUID, err := s.model.GetAddressOwnerByUUID(ctx, id)
	if err != nil {
		return nil, err
	}
	return userUUID, nil
}

// GetAddresses gets a slice of addresses for a given user
func (s *Service) GetAddresses(ctx context.Context, userID string) ([]*Address, error) {
	contextLogger := log.WithContext(ctx)
	contextLogger.Infof("service: GetAddresses(ctx, userID=%q) started", userID)

	al, err := s.model.GetAddresses(ctx, userID)
	if err == postgres.ErrUserNotFound {
		return nil, ErrUserNotFound
	}
	if err != nil {
		return nil, errors.Wrapf(err, "service: s.model.GetAddresses(ctx, userID=%q)", userID)
	}

	addresses := make([]*Address, 0, 32)
	for _, a := range al {
		addr := Address{
			Object:      "address",
			ID:          a.UUID,
			UserID:      a.UsrUUID,
			Typ:         a.Typ,
			ContactName: a.ContactName,
			Addr1:       a.Addr1,
			Addr2:       a.Addr2,
			City:        a.City,
			County:      a.County,
			Postcode:    a.Postcode,
			CountryCode: a.CountryCode,
			Created:     a.Created,
			Modified:    a.Modified,
		}
		addresses = append(addresses, &addr)
	}
	return addresses, nil
}

// PartialUpdateAddress updates one or more fields of the given
// address record returing the updates address.
func (s *Service) PartialUpdateAddress(ctx context.Context, addressID string, typ, contactName, addr1, addr2, city, county, postcode, countryCode *string) (*Address, error) {
	contextLogger := log.WithContext(ctx)
	contextLogger.Infof("service: PartialUpdateAddress(ctx, addressID=%q, ...) started", addressID)

	a, err := s.model.PartialUpdateAddress(ctx, addressID, typ, contactName,
		addr1, addr2, city, county, postcode, countryCode)
	if err == postgres.ErrAddressNotFound {
		return nil, ErrAddressNotFound
	}
	if err != nil {
		return nil, errors.Wrapf(err,
			"service: s.model.PartialUpdateAddress(ctx, addressID=%q, typ=%v)",
			addressID, typ)
	}
	addr := Address{
		Object:      "address",
		ID:          a.UUID,
		UserID:      a.UsrUUID,
		Typ:         a.Typ,
		ContactName: a.ContactName,
		Addr1:       a.Addr1,
		Addr2:       a.Addr2,
		City:        a.City,
		County:      a.County,
		Postcode:    a.Postcode,
		CountryCode: a.CountryCode,
		Created:     a.Created,
		Modified:    a.Modified,
	}

	if err := s.PublishTopicEvent(ctx, EventAddressUpdated, &addr); err != nil {
		return nil, errors.Wrapf(err,
			"service: s.PublishTopicEvent(ctx, event=%q, data=%v) failed",
			EventAddressUpdated, addr)
	}
	contextLogger.WithFields(log.Fields{
		"Object":      "address",
		"ID":          a.UUID,
		"UserID":      a.UsrUUID,
		"Typ":         a.Typ,
		"ContactName": a.ContactName,
		"Addr1":       a.Addr1,
		"Addr2":       a.Addr2,
		"City":        a.City,
		"County":      a.County,
		"Postcode":    a.Postcode,
		"CountryCode": a.CountryCode,
		"Created":     a.Created,
		"Modified":    a.Modified,
	}).Infof("service: EventAddressUpdated published")
	return &addr, nil
}

// DeleteAddress deletes an address by ID.
func (s *Service) DeleteAddress(ctx context.Context, addrUUID string) error {
	err := s.model.DeleteAddressByUUID(ctx, addrUUID)
	if err == postgres.ErrAddressNotFound {
		return ErrAddressNotFound
	}
	if err != nil {
		return errors.Wrap(err, "service: s.model.DeleteAddressByUUID(ctx, addressUUID=%q) failed")
	}
	return nil
}
