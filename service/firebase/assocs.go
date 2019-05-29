package firebase

import (
	"context"
	"time"

	"github.com/pkg/errors"
)

// CatalogProductAssoc maps products to leaf nodes in the catalogue hierarchy
type CatalogProductAssoc struct {
	Path     string    `json:"path"`
	SKU      string    `json:"sku"`
	Pri      int       `json:"pri"`
	Created  time.Time `json:"created"`
	Modified time.Time `json:"modified"`
}

// BatchCreateCatalogProductAssocs creates a set of catalog product
// associations either completing with all or failing with none
// being added.
func (s *Service) BatchCreateCatalogProductAssocs(ctx context.Context, cpas map[string][]string) error {
	err := s.model.BatchCreateCatalogProductAssocs(ctx, cpas)
	if err != nil {
		return errors.Wrap(err, "service: BatchCreateCatalogProductAssocs")
	}
	return nil
}

// CreateCatalogProductAssocs associates an existing product to a catalog entry.
func (s *Service) CreateCatalogProductAssocs(ctx context.Context, path, sku string) (*CatalogProductAssoc, error) {
	cpa, err := s.model.CreateCatalogProductAssoc(ctx, path, sku)
	if err != nil {
		return nil, errors.Wrapf(err, "service: create catalog product assoc sku=%q", sku)
	}
	scpa := CatalogProductAssoc{
		Path:     cpa.Path,
		SKU:      cpa.SKU,
		Pri:      cpa.Pri,
		Created:  cpa.Created,
		Modified: cpa.Modified,
	}
	return &scpa, nil
}

// HasCatalogProductAssocs returns true if any catalog product associations
// exist.
func (s *Service) HasCatalogProductAssocs(ctx context.Context) (bool, error) {
	has, err := s.model.HasCatalogProductAssocs(ctx)
	if err != nil {
		return false, errors.Wrap(err, "service: has catalog product assocs")
	}
	return has, nil
}

// A SKU handles unique product SKUs.
type SKU string

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

// GetCatalogAssocs returns the catalog product associations
func (s *Service) GetCatalogAssocs(ctx context.Context) (map[string]*Assoc, error) {
	cpo, err := s.model.GetCatalogProductAssocs(ctx)
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

// UpdateCatalogProductAssocs updates the catalog product associations
// func (s *Service) UpdateCatalogProductAssocs(ctx context.Context, cpo []*postgres.catalogProductAssoc) error {
// 	err := s.model.UpdateCatalogProductAssocs(ctx, cpo)
// 	if err != nil {
// 		return err
// 	}
// 	return nil
// }

// DeleteCatalogAssocs delete all catalog product associations.
func (s *Service) DeleteCatalogAssocs(ctx context.Context) (affected int64, err error) {
	n, err := s.model.DeleteCatalogProductAssocs(ctx)
	if err != nil {
		return 0, errors.Wrapf(err, "service: delete catalog assocs")
	}
	return n, nil
}
