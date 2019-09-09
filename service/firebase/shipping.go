package firebase

import (
	"context"
	"time"

	"bitbucket.org/andyfusniakteam/ecom-api-go/model/postgres"
	"github.com/pkg/errors"
)

// ErrShippingTarrifCodeExists error for duplicates.
var ErrShippingTarrifCodeExists = errors.New("service: shipping tarrif code exists")

// ErrShippingTarrifNotFound error
var ErrShippingTarrifNotFound = errors.New("service: shipping tarrif not found")

// ShippingTarrif holds a single shipping tarrif.
type ShippingTarrif struct {
	Object       string    `json:"object"`
	ID           string    `json:"id"`
	CountryCode  string    `json:"country_code"`
	ShippingCode string    `json:"shipping_code"`
	Name         string    `json:"name"`
	Price        int       `json:"price"`
	TaxCode      string    `json:"tax_code"`
	Created      time.Time `json:"created"`
	Modified     time.Time `json:"modified"`
}

// CreateShippingTarrif creates a new shipping tarrif entry.
func (s *Service) CreateShippingTarrif(ctx context.Context, countryCode, shippingCode, name string, price int, taxCode string) (*ShippingTarrif, error) {
	ptarrif, err := s.model.CreateShippingTarrif(ctx, countryCode, shippingCode, name, price, taxCode)
	if err != nil {
		if err == postgres.ErrShippingTarrifCodeExists {
			return nil, ErrShippingTarrifCodeExists
		}
		return nil, errors.Wrapf(err, "s.model.CreateShippingTarrif(ctx, countryCode=%q, shippingCode=%q, name=%q, price=%d, taxCode=%q) failed", countryCode, shippingCode, name, price, taxCode)
	}

	shippingTarrif := ShippingTarrif{
		Object:       "shipping_tarrif",
		ID:           ptarrif.UUID,
		CountryCode:  ptarrif.CountryCode,
		ShippingCode: ptarrif.ShippingCode,
		Name:         ptarrif.Name,
		Price:        ptarrif.Price,
		TaxCode:      ptarrif.TaxCode,
	}
	return &shippingTarrif, nil
}

// GetShippingTarrif returns a ShippingTarrif by ID
func (s *Service) GetShippingTarrif(ctx context.Context, shippingTarrifID string) (*ShippingTarrif, error) {
	ptarrif, err := s.model.GetShippingTarrifByUUID(ctx, shippingTarrifID)
	if err != nil {
		if err == postgres.ErrShippingTarrifCodeExists {
			return nil, ErrShippingTarrifCodeExists
		}
		return nil, errors.Wrapf(err, "s.model.GetShippingTarrif(ctx, shippingTarrifID=%q)", shippingTarrifID)
	}

	shippingTarrif := ShippingTarrif{
		Object:       "shipping_tarrif",
		ID:           ptarrif.UUID,
		CountryCode:  ptarrif.CountryCode,
		ShippingCode: ptarrif.ShippingCode,
		Name:         ptarrif.Name,
		Price:        ptarrif.Price,
		TaxCode:      ptarrif.TaxCode,
		Created:      ptarrif.Created,
		Modified:     ptarrif.Modified,
	}
	return &shippingTarrif, nil
}

// GetShippingTarrifs returns a list of ShippingTarrifs.
func (s *Service) GetShippingTarrifs(ctx context.Context) ([]*ShippingTarrif, error) {
	ptarrifs, err := s.model.GetShippingTarrifs(ctx)
	if err != nil {
		return nil, errors.Wrapf(err, "s.model.GetShippingTarrifs(ctx) failed")
	}

	shippingTarrifList := make([]*ShippingTarrif, 0, len(ptarrifs))
	for _, s := range ptarrifs {
		shippingTarrif := ShippingTarrif{
			Object:       "shipping_tarrif",
			ID:           s.UUID,
			CountryCode:  s.CountryCode,
			ShippingCode: s.ShippingCode,
			Name:         s.Name,
			Price:        s.Price,
			TaxCode:      s.TaxCode,
			Created:      s.Created,
			Modified:     s.Modified,
		}
		shippingTarrifList = append(shippingTarrifList, &shippingTarrif)
	}
	return shippingTarrifList, nil
}

// UpdateShippingTarrif updates a shipping tarrif.
func (s *Service) UpdateShippingTarrif(ctx context.Context, shoppingTarrifID, countryCode, shippingCode, name string, price int, taxCode string) (*ShippingTarrif, error) {
	ptarrif, err := s.model.UpdateShippingTarrif(ctx, shoppingTarrifID, countryCode, shippingCode, name, price, taxCode)
	if err != nil {
		if err == postgres.ErrShippingTarrifNotFound {
			return nil, ErrShippingTarrifNotFound
		} else if err == postgres.ErrShippingTarrifCodeExists {
			return nil, ErrShippingTarrifCodeExists
		}
		return nil, errors.Wrapf(err, ".model.UpdateShippingTarrif(ctx, shoppingTarrifID=%q, countryCode=%q, shippingcode=%q, name=%q, price=%d, taxCode=%q)", shoppingTarrifID, countryCode, shoppingTarrifID, name, price, taxCode)
	}

	shippingTarrif := ShippingTarrif{
		Object:       "shipping_tarrif",
		ID:           ptarrif.UUID,
		CountryCode:  ptarrif.CountryCode,
		ShippingCode: ptarrif.ShippingCode,
		Name:         ptarrif.Name,
		Price:        ptarrif.Price,
		TaxCode:      ptarrif.TaxCode,
		Created:      ptarrif.Created,
		Modified:     ptarrif.Modified,
	}
	return &shippingTarrif, nil
}

// DeleteShippingTarrif delete a shipping tarrif.
func (s *Service) DeleteShippingTarrif(ctx context.Context, shippingTarrifID string) error {
	err := s.model.DeleteShippingTarrifByUUID(ctx, shippingTarrifID)
	if err != nil {
		if err == postgres.ErrShippingTarrifNotFound {
			return ErrShippingTarrifNotFound
		}
		return errors.Wrapf(err, "s.model.DeleteShippingTarrifByUUID(ctx, shippingTarrifUUID=%q)", shippingTarrifID)
	}
	return nil
}
