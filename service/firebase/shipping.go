package firebase

import (
	"context"
	"time"

	"bitbucket.org/andyfusniakteam/ecom-api-go/model/postgres"
	"github.com/pkg/errors"
)

// ErrShippingTariffCodeExists error for duplicates.
var ErrShippingTariffCodeExists = errors.New("service: shipping tariff code exists")

// ErrShippingTariffNotFound error
var ErrShippingTariffNotFound = errors.New("service: shipping tariff not found")

// ShippingTariff holds a single shipping tariff.
type ShippingTariff struct {
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

// CreateShippingTariff creates a new shipping tariff entry.
func (s *Service) CreateShippingTariff(ctx context.Context, countryCode, shippingCode, name string, price int, taxCode string) (*ShippingTariff, error) {
	ptariff, err := s.model.CreateShippingTariff(ctx, countryCode, shippingCode, name, price, taxCode)
	if err != nil {
		if err == postgres.ErrShippingTariffCodeExists {
			return nil, ErrShippingTariffCodeExists
		}
		return nil, errors.Wrapf(err, "s.model.CreateShippingTariff(ctx, countryCode=%q, shippingCode=%q, name=%q, price=%d, taxCode=%q) failed", countryCode, shippingCode, name, price, taxCode)
	}

	shippingTariff := ShippingTariff{
		Object:       "shipping_tariff",
		ID:           ptariff.UUID,
		CountryCode:  ptariff.CountryCode,
		ShippingCode: ptariff.ShippingCode,
		Name:         ptariff.Name,
		Price:        ptariff.Price,
		TaxCode:      ptariff.TaxCode,
	}
	return &shippingTariff, nil
}

// GetShippingTariff returns a ShippingTariff by ID
func (s *Service) GetShippingTariff(ctx context.Context, shippingTariffID string) (*ShippingTariff, error) {
	ptariff, err := s.model.GetShippingTariffByUUID(ctx, shippingTariffID)
	if err != nil {
		if err == postgres.ErrShippingTariffCodeExists {
			return nil, ErrShippingTariffCodeExists
		}
		return nil, errors.Wrapf(err, "s.model.GetShippingTariff(ctx, shippingTariffID=%q)", shippingTariffID)
	}

	shippingTariff := ShippingTariff{
		Object:       "shipping_tariff",
		ID:           ptariff.UUID,
		CountryCode:  ptariff.CountryCode,
		ShippingCode: ptariff.ShippingCode,
		Name:         ptariff.Name,
		Price:        ptariff.Price,
		TaxCode:      ptariff.TaxCode,
		Created:      ptariff.Created,
		Modified:     ptariff.Modified,
	}
	return &shippingTariff, nil
}

// GetShippingTariffs returns a list of ShippingTariffs.
func (s *Service) GetShippingTariffs(ctx context.Context) ([]*ShippingTariff, error) {
	ptariffs, err := s.model.GetShippingTariffs(ctx)
	if err != nil {
		return nil, errors.Wrapf(err, "s.model.GetShippingTariffs(ctx) failed")
	}

	shippingTariffList := make([]*ShippingTariff, 0, len(ptariffs))
	for _, s := range ptariffs {
		shippingTariff := ShippingTariff{
			Object:       "shipping_tariff",
			ID:           s.UUID,
			CountryCode:  s.CountryCode,
			ShippingCode: s.ShippingCode,
			Name:         s.Name,
			Price:        s.Price,
			TaxCode:      s.TaxCode,
			Created:      s.Created,
			Modified:     s.Modified,
		}
		shippingTariffList = append(shippingTariffList, &shippingTariff)
	}
	return shippingTariffList, nil
}

// UpdateShippingTariff updates a shipping tariff.
func (s *Service) UpdateShippingTariff(ctx context.Context, shoppingTariffID, countryCode, shippingCode, name string, price int, taxCode string) (*ShippingTariff, error) {
	ptariff, err := s.model.UpdateShippingTariff(ctx, shoppingTariffID, countryCode, shippingCode, name, price, taxCode)
	if err != nil {
		if err == postgres.ErrShippingTariffNotFound {
			return nil, ErrShippingTariffNotFound
		} else if err == postgres.ErrShippingTariffCodeExists {
			return nil, ErrShippingTariffCodeExists
		}
		return nil, errors.Wrapf(err, ".model.UpdateShippingTariff(ctx, shoppingTariffID=%q, countryCode=%q, shippingcode=%q, name=%q, price=%d, taxCode=%q)", shoppingTariffID, countryCode, shoppingTariffID, name, price, taxCode)
	}

	shippingTariff := ShippingTariff{
		Object:       "shipping_tariff",
		ID:           ptariff.UUID,
		CountryCode:  ptariff.CountryCode,
		ShippingCode: ptariff.ShippingCode,
		Name:         ptariff.Name,
		Price:        ptariff.Price,
		TaxCode:      ptariff.TaxCode,
		Created:      ptariff.Created,
		Modified:     ptariff.Modified,
	}
	return &shippingTariff, nil
}

// DeleteShippingTariff delete a shipping tariff.
func (s *Service) DeleteShippingTariff(ctx context.Context, shippingTariffID string) error {
	err := s.model.DeleteShippingTariffByUUID(ctx, shippingTariffID)
	if err != nil {
		if err == postgres.ErrShippingTariffNotFound {
			return ErrShippingTariffNotFound
		}
		return errors.Wrapf(err, "s.model.DeleteShippingTariffByUUID(ctx, shippingTariffUUID=%q)", shippingTariffID)
	}
	return nil
}
