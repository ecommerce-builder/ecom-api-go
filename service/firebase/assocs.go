package firebase

import (
	"context"
	"fmt"
	"time"

	"bitbucket.org/andyfusniakteam/ecom-api-go/model/postgres"
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

// ProductsCategories response body
type ProductsCategories struct {
	Object       string    `json:"object"`
	ID           string    `json:"id"`
	ProductID    string    `json:"product_id"`
	ProductPath  string    `json:"product_path"`
	ProductSKU   string    `json:"product_sku"`
	ProductName  string    `json:"product_name"`
	CategoryID   string    `json:"category_id"`
	CategoryPath string    `json:"category_path"`
	Created      time.Time `json:"created"`
	Modified     time.Time `json:"modified"`
}

// CreateProductsCategories request
type CreateProductsCategories struct {
	ProductID  string `json:"product_id"`
	CategoryID string `json:"category_id"`
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
// func (s *Service) CreateCategoryProductAssoc(ctx context.Context, path, sku string) (*CategoryProductAssoc, error) {
// 	cpa, err := s.model.CreateCategoryProductAssoc(ctx, path, sku)
// 	if err != nil {
// 		return nil, errors.Wrapf(err, "service: create catalog product assoc sku=%q", sku)
// 	}
// 	scpa := CategoryProductAssoc{
// 		Path:     cpa.Path,
// 		SKU:      cpa.SKU,
// 		Pri:      cpa.Pri,
// 		Created:  cpa.Created,
// 		Modified: cpa.Modified,
// 	}
// 	return &scpa, nil
// }

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
	Products *ProductList `json:"products"`
}

// GetProductsCategoriesList returns a list of product to category associations.
func (s *Service) GetProductsCategoriesList(ctx context.Context) ([]*ProductsCategories, error) {
	cps, err := s.model.GetProductsCategories(ctx)
	if err != nil {
		return nil, errors.Wrapf(err, "service: s.model.GetCategoryProductAssocs(ctx) failed")
	}

	productsCategories := make([]*ProductsCategories, 0, len(cps))
	for _, pc := range cps {
		pc := ProductsCategories{
			Object:       "products_category",
			ID:           pc.UUID,
			ProductID:    pc.ProductUUID,
			ProductPath:  pc.ProductPath,
			ProductSKU:   pc.ProductSKU,
			ProductName:  pc.ProductName,
			CategoryID:   pc.CategoryUUID,
			CategoryPath: pc.CategoryPath,
			Created:      pc.Created,
			Modified:     pc.Modified,
		}
		productsCategories = append(productsCategories, &pc)
	}

	return productsCategories, nil
}

// UpdateProductsCategories batch updates products categories associations
func (s *Service) UpdateProductsCategories(ctx context.Context, cpcs []*CreateProductsCategories) ([]*ProductsCategories, error) {
	createProductsCategories := make([]*postgres.CreateProductsCategoriesRow, 0, len(cpcs))
	for _, cp := range cpcs {
		c := postgres.CreateProductsCategoriesRow{
			ProductUUID:  cp.ProductID,
			CategoryUUID: cp.CategoryID,
		}
		createProductsCategories = append(createProductsCategories, &c)
	}

	list, err := s.model.UpdateProductsCategories(ctx, createProductsCategories)
	if err != nil {
		if err == postgres.ErrProductNotFound {
			return nil, ErrProductNotFound
		} else if err == postgres.ErrCategoryNotFound {
			return nil, ErrCategoryNotFound
		}
		return nil, errors.Wrap(err, "s.model.UpdateProductsCategories(ctx, createProductsCategories) failed")
	}

	fmt.Printf("%#v\n", len(list))
	results := make([]*ProductsCategories, 0, len(list))
	for _, l := range list {
		pc := ProductsCategories{
			Object:       "products_categories",
			ID:           l.UUID,
			ProductID:    l.ProductUUID,
			ProductPath:  l.ProductPath,
			ProductSKU:   l.ProductSKU,
			ProductName:  l.ProductName,
			CategoryID:   l.CategoryUUID,
			CategoryPath: l.CategoryPath,
			Created:      l.Created,
			Modified:     l.Modified,
		}
		results = append(results, &pc)
	}
	return results, nil
}

// GetCategoryProductAssocs returns all of the category product associations
// keyed by key. key has a value of either `id` or `path`
func (s *Service) GetCategoryProductAssocs(ctx context.Context, key string) (map[string]*Assoc, error) {
	cpo, err := s.model.GetCategoryProductAssocsFull(ctx)
	if err != nil {
		return nil, err
	}
	assocs := make(map[string]*Assoc)
	for _, v := range cpo {
		var k string
		if key == "id" {
			k = v.CategoryUUID
		} else {
			k = v.CategoryPath
		}
		if _, ok := assocs[k]; !ok {
			assocs[k] = &Assoc{
				Products: &ProductList{
					Object: "list",
					Data:   make([]*Product, 0),
				},
			}
		}
		p := Product{
			Object:   "product",
			ID:       v.ProductUUID,
			Path:     v.ProductPath,
			SKU:      v.ProductSKU,
			Name:     v.ProductName,
			Created:  v.ProductCreated,
			Modified: v.ProductModified,
		}
		assocs[k].Products.Data = append(assocs[k].Products.Data, &p)
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

// PurgeProductsCategories delete all catalog product associations.
func (s *Service) PurgeProductsCategories(ctx context.Context) error {
	err := s.model.PurgeProductsCategories(ctx)
	if err != nil {
		return errors.Wrapf(err, "service: purge products categories")
	}
	return nil
}

// DeleteCategoryAssocs delete all catalog product associations.
func (s *Service) DeleteCategoryAssocs(ctx context.Context) (affected int64, err error) {
	n, err := s.model.DeleteCategoryProductAssocs(ctx)
	if err != nil {
		return 0, errors.Wrapf(err, "service: delete catalog assocs")
	}
	return n, nil
}
