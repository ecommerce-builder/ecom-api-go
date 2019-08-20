package firebase

import (
	"context"
	"time"

	"bitbucket.org/andyfusniakteam/ecom-api-go/model/postgres"
	"github.com/pkg/errors"
)

// ErrPriceListNotFound is where no tier pricing could be found
// for the given sku and tier ref.
var ErrPriceListNotFound = errors.New("service: price list not found")

// ErrPriceListCodeTaken error
var ErrPriceListCodeTaken = errors.New("service: price list code already taken")

// ErrPriceListInUse error
var ErrPriceListInUse = errors.New("service: price list has associated prices")

// PriceList represents a price list.
type PriceList struct {
	Object        string    `json:"object"`
	ID            string    `json:"id"`
	PriceListCode string    `json:"price_list_code"`
	Name          string    `json:"name"`
	Description   string    `json:"description"`
	Created       time.Time `json:"created"`
	Modified      time.Time `json:"modified"`
}

// PriceListCreate request body for creating a new price list.
type PriceListCreate struct {
	PriceListCode string `json:"price_list_code"`
	Name          string `json:"name"`
	Description   string `json:"description"`
}

// CreatePriceList creates a new price list returning the newly created price list.
func (s *Service) CreatePriceList(ctx context.Context, p *PriceListCreate) (*PriceList, error) {
	row, err := s.model.CreatePriceList(ctx, p.PriceListCode, p.Name, p.Description)
	if err != nil {
		if err == postgres.ErrPriceListCodeTaken {
			return nil, ErrPriceListCodeTaken
		}
		return nil, errors.Wrapf(err, "service: s.model.CreatePriceList(ctx, code=%q, name=%q, description=%q) failed", p.PriceListCode, p.Name, p.Description)
	}
	priceList := PriceList{
		Object:        "price_list",
		ID:            row.UUID,
		PriceListCode: row.Code,
		Name:          row.Name,
		Description:   row.Description,
		Created:       row.Created,
		Modified:      row.Modified,
	}
	return &priceList, nil
}

// GetPriceList returns a single price list.
func (s *Service) GetPriceList(ctx context.Context, priceListID string) (*PriceList, error) {
	row, err := s.model.GetPriceList(ctx, priceListID)
	if err != nil {
		if err == postgres.ErrPriceListNotFound {
			return nil, ErrPriceListNotFound
		}
		return nil, errors.Wrapf(err, "s.model.GetPriceList(ctx, priceListID=%q)", priceListID)
	}
	priceList := PriceList{
		Object:        "price_list",
		ID:            row.UUID,
		PriceListCode: row.Code,
		Name:          row.Name,
		Description:   row.Description,
		Created:       row.Created,
		Modified:      row.Modified,
	}
	return &priceList, nil
}

// GetPriceLists returns a list of PriceLists.
func (s *Service) GetPriceLists(ctx context.Context) ([]*PriceList, error) {
	rows, err := s.model.GetPriceLists(ctx)
	if err != nil {
		if err == postgres.ErrPriceListNotFound {
			return nil, ErrPriceListNotFound
		}
		return nil, errors.Wrapf(err, "s.model.GetPriceLists(ctx) failed")
	}

	priceLists := make([]*PriceList, 0, len(rows))
	for _, t := range rows {
		pl := PriceList{
			Object:        "price_list",
			ID:            t.UUID,
			PriceListCode: t.Code,
			Name:          t.Name,
			Description:   t.Description,
			Created:       t.Created,
			Modified:      t.Modified,
		}
		priceLists = append(priceLists, &pl)
	}
	return priceLists, nil
}

// UpdatePriceList updates a price list with a new price list code, name
// and description.
func (s *Service) UpdatePriceList(ctx context.Context, priceListID string, p *PriceListCreate) (*PriceList, error) {
	row, err := s.model.UpdatePriceList(ctx, priceListID, p.PriceListCode, p.Name, p.Description)
	if err != nil {
		if err == postgres.ErrPriceListNotFound {
			return nil, ErrPriceListNotFound
		} else if err == postgres.ErrPriceListCodeTaken {
			return nil, ErrPriceListCodeTaken
		}
		return nil, errors.Wrapf(err, "service: s.model.UpdatePriceList(ctx, priceListID=%q, code=%q, name=%q, description=%q) failed", priceListID, p.PriceListCode, p.Name, p.Description)
	}
	priceList := PriceList{
		Object:        "price_list",
		ID:            row.UUID,
		PriceListCode: row.Code,
		Name:          row.Name,
		Description:   row.Description,
		Created:       row.Created,
		Modified:      row.Modified,
	}
	return &priceList, nil
}

// DeletePriceList deletes a single price list.
func (s *Service) DeletePriceList(ctx context.Context, priceListID string) error {
	err := s.model.DeletePriceList(ctx, priceListID)
	if err != nil {
		if err == postgres.ErrPriceListNotFound {
			return ErrPriceListNotFound
		} else if err == postgres.ErrPriceListInUse {
			return ErrPriceListInUse
		}
		return errors.Wrapf(err, "s.model.DeletePriceList(ctx, pricingListID=%q) failed", priceListID)
	}
	return nil
}
