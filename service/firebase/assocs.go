package firebase

import (
	"context"
	"time"

	"github.com/pkg/errors"
)

// CategoryProductAssoc maps products to leaf nodes in the catalogue hierarchy
type CategoryProductAssoc struct {
	Path     string    `json:"path"`
	SKU      string    `json:"sku"`
	Pri      int       `json:"pri"`
	Created  time.Time `json:"created"`
	Modified time.Time `json:"modified"`
}

// CreateCategoryProductAssocs creates a set of catalog product
// associations either completing with all or failing with none
// being added.
func (s *Service) CreateCategoryProductAssocs(ctx context.Context, cpas map[string][]string) error {
	err := s.model.CreateCategoryProductAssocs(ctx, cpas)
	if err != nil {
		return errors.Wrap(err, "service: CreateCategoryProductAssocs")
	}
	return nil
}

// CreateCategoryProductAssoc associates an existing product to a catalog entry.
func (s *Service) CreateCategoryProductAssoc(ctx context.Context, path, sku string) (*CategoryProductAssoc, error) {
	cpa, err := s.model.CreateCategoryProductAssoc(ctx, path, sku)
	if err != nil {
		return nil, errors.Wrapf(err, "service: create catalog product assoc sku=%q", sku)
	}
	scpa := CategoryProductAssoc{
		Path:     cpa.Path,
		SKU:      cpa.SKU,
		Pri:      cpa.Pri,
		Created:  cpa.Created,
		Modified: cpa.Modified,
	}
	return &scpa, nil
}

// HasCategoryProductAssocs returns true if any catalog product associations
// exist.
func (s *Service) HasCategoryProductAssocs(ctx context.Context) (bool, error) {
	has, err := s.model.HasCategoryProductAssocs(ctx)
	if err != nil {
		return false, errors.Wrap(err, "service: has catalog product assocs")
	}
	return has, nil
}

// An AssocProduct holds details of a product in the context of an AssocSet.
type AssocProduct struct {
	SKU      string    `json:"sku"`
	Created  time.Time `json:"created"`
	Modified time.Time `json:"modified"`
}

// Assoc details a catalog association including products.
type Assoc struct {
	Path     string         `json:"path"`
	Products []AssocProduct `json:"products"`
}

// GetCategoryAssocs returns all of the category product associations
func (s *Service) GetCategoryAssocs(ctx context.Context) (map[string]*Assoc, error) {
	cpo, err := s.model.GetCategoryProductAssocs(ctx)
	if err != nil {
		return nil, err
	}
	assocs := make(map[string]*Assoc)
	for _, v := range cpo {
		if _, ok := assocs[v.Path]; !ok {
			assocs[v.Path] = &Assoc{
				Path:     v.Path,
				Products: make([]AssocProduct, 0),
			}
		}
		p := AssocProduct{
			SKU:      v.SKU,
			Created:  v.Created,
			Modified: v.Modified,
		}
		assocs[v.Path].Products = append(assocs[v.Path].Products, p)
	}
	return assocs, nil
}

// UpdateCategoryProductAssocs updates the category product associations
// func (s *Service) UpdateCategoryProductAssocs(ctx context.Context, cpo []*postgres.catalogProductAssoc) error {
// 	err := s.model.UpdateCategoryProductAssocs(ctx, cpo)
// 	if err != nil {
// 		return err
// 	}
// 	return nil
// }

// DeleteCategoryAssocs delete all catalog product associations.
func (s *Service) DeleteCategoryAssocs(ctx context.Context) (affected int64, err error) {
	n, err := s.model.DeleteCategoryProductAssocs(ctx)
	if err != nil {
		return 0, errors.Wrapf(err, "service: delete catalog assocs")
	}
	return n, nil
}
