package firebase

import (
	"context"
	"time"
)

// Address contains address information for a Customer
type Address struct {
	AddrUUID    string    `json:"uuid"`
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

// CreateAddress creates a new address for a customer
func (s *Service) CreateAddress(ctx context.Context, customerUUID, typ, contactName, addr1 string, addr2 *string, city string, county *string, postcode string, country string) (*Address, error) {
	customerID, err := s.model.GetCustomerIDByUUID(ctx, customerUUID)
	if err != nil {
		return nil, err
	}

	a, err := s.model.CreateAddress(ctx, customerID, typ, contactName, addr1, addr2, city, county, postcode, country)
	if err != nil {
		return nil, err
	}

	aa := Address{
		AddrUUID:    a.AddrUUID,
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

// GetAddress gets an address by UUID
func (s *Service) GetAddress(ctx context.Context, uuid string) (*Address, error) {
	addr, err := s.model.GetAddressByUUID(ctx, uuid)
	if err != nil {
		if s.model.IsNotExist(err) {
			if ne, ok := err.(*ResourceError); ok {
				return nil, &ResourceError{
					Op:       "GetAddress",
					Resource: "address",
					UUID:     uuid,
					Err:      ne.Err,
				}
			}
		}
		return nil, err
	}

	aa := Address{
		AddrUUID:    addr.AddrUUID,
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

// GetAddressOwner returns the Customer that owns the address with the given UUID.
func (s *Service) GetAddressOwner(ctx context.Context, uuid string) (*string, error) {
	customerUUID, err := s.model.GetAddressOwnerByUUID(ctx, uuid)
	if err != nil {
		return nil, err
	}
	return customerUUID, nil
}

// GetAddresses gets a slice of addresses for a given customer
func (s *Service) GetAddresses(ctx context.Context, customerUUID string) ([]*Address, error) {
	customerID, err := s.model.GetCustomerIDByUUID(ctx, customerUUID)
	if err != nil {
		return nil, err
	}

	al, err := s.model.GetAddresses(ctx, customerID)
	if err != nil {
		return nil, err
	}

	results := make([]*Address, 0, 32)
	for _, v := range al {
		i := Address{
			AddrUUID:    v.AddrUUID,
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

// DeleteAddress deletes an address by uuid
func (s *Service) DeleteAddress(ctx context.Context, addrUUID string) error {
	err := s.model.DeleteAddressByUUID(ctx, addrUUID)
	if err != nil {
		return err
	}

	return nil
}

