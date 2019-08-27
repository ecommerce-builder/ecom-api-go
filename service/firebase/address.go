package firebase

import (
	"context"
	"time"
)

// Address contains address information for a user
type Address struct {
	ID          string    `json:"id"`
	Typ         string    `json:"typ"`
	ContactName string    `json:"contact_name"`
	Addr1       string    `json:"addr1"`
	Addr2       *string   `json:"addr2,omitempty"`
	City        string    `json:"city"`
	County      *string   `json:"county,omitempty"`
	Postcode    string    `json:"postcode"`
	Country     string    `json:"country"`
	Created     time.Time `json:"created"`
	Modified    time.Time `json:"modified"`
}

// CreateAddress creates a new address for a user
func (s *Service) CreateAddress(ctx context.Context, userUUID, typ, contactName, addr1 string, addr2 *string, city string, county *string, postcode string, country string) (*Address, error) {
	userID, err := s.model.GetUserIDByUUID(ctx, userUUID)
	if err != nil {
		return nil, err
	}

	a, err := s.model.CreateAddress(ctx, userID, typ, contactName, addr1, addr2, city, county, postcode, country)
	if err != nil {
		return nil, err
	}

	aa := Address{
		ID:          a.UUID,
		Typ:         a.Typ,
		ContactName: a.ContactName,
		Addr1:       a.Addr1,
		Addr2:       a.Addr2,
		City:        a.City,
		County:      a.County,
		Postcode:    a.Postcode,
		Country:     a.Country,
		Created:     a.Created,
		Modified:    a.Modified,
	}
	return &aa, nil
}

// GetAddress gets an address by ID.
func (s *Service) GetAddress(ctx context.Context, id string) (*Address, error) {
	addr, err := s.model.GetAddressByUUID(ctx, id)
	if err != nil {
		if s.model.IsNotExist(err) {
			if ne, ok := err.(*ResourceError); ok {
				return nil, &ResourceError{
					Op:       "GetAddress",
					Resource: "address",
					ID:       id,
					Err:      ne.Err,
				}
			}
		}
		return nil, err
	}

	aa := Address{
		ID:          addr.UUID,
		Typ:         addr.Typ,
		ContactName: addr.ContactName,
		Addr1:       addr.Addr1,
		Addr2:       addr.Addr2,
		City:        addr.City,
		County:      addr.County,
		Postcode:    addr.Postcode,
		Country:     addr.Country,
		Created:     addr.Created,
		Modified:    addr.Modified,
	}
	return &aa, nil
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
func (s *Service) GetAddresses(ctx context.Context, userUUID string) ([]*Address, error) {
	userID, err := s.model.GetUserIDByUUID(ctx, userUUID)
	if err != nil {
		return nil, err
	}

	al, err := s.model.GetAddresses(ctx, userID)
	if err != nil {
		return nil, err
	}

	results := make([]*Address, 0, 32)
	for _, v := range al {
		i := Address{
			ID:          v.UUID,
			Typ:         v.Typ,
			ContactName: v.ContactName,
			Addr1:       v.Addr1,
			Addr2:       v.Addr2,
			City:        v.City,
			County:      v.County,
			Postcode:    v.Postcode,
			Country:     v.Country,
			Created:     v.Created,
			Modified:    v.Modified,
		}
		results = append(results, &i)
	}

	return results, nil
}

// DeleteAddress deletes an address by ID.
func (s *Service) DeleteAddress(ctx context.Context, addrUUID string) error {
	err := s.model.DeleteAddressByUUID(ctx, addrUUID)
	if err != nil {
		return err
	}

	return nil
}
