package firebase

import (
	"context"
	"fmt"
	"time"

	"bitbucket.org/andyfusniakteam/ecom-api-go/model/postgres"
	"github.com/pkg/errors"
)

// ErrProductCategoryExists error
var ErrProductCategoryExists = errors.New("service: product to category association exists")

// ErrProductCategoryNotFound error
var ErrProductCategoryNotFound = errors.New("service: product to category association not found")

// ProductCategoryRequestBody request used for linking a product to a category.
type ProductCategoryRequestBody struct {
	CategoryID string `json:"category_id"`
	ProductID  string `json:"product_id"`
}

// ProductCategory represents an association between a product and a category.
type ProductCategory struct {
	Object     string    `json:"object"`
	ID         string    `json:"id"`
	ProductID  string    `json:"product_id"`
	CategoryID string    `json:"category_id"`
	Pri        int       `json:"pri"`
	Created    time.Time `json:"created"`
	Modified   time.Time `json:"modified"`
}

// ProductCategoryAssoc maps products to leaf nodes in the catalogue hierarchy
type ProductCategoryAssoc struct {
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

// AddProductCategory associates a product to a leaf category
func (s *Service) AddProductCategory(ctx context.Context, request *ProductCategoryRequestBody) (*ProductCategory, error) {
	row, err := s.model.AddProductCategory(ctx, request.CategoryID, request.ProductID)
	if err != nil {
		if err == postgres.ErrCategoryNotFound {
			return nil, ErrCategoryNotFound
		} else if err == postgres.ErrCategoryNotLeaf {
			return nil, ErrCategoryNotLeaf
		} else if err == postgres.ErrProductNotFound {
			return nil, ErrProductNotFound
		} else if err == postgres.ErrProductCategoryExists {
			return nil, ErrProductCategoryExists
		}
		return nil, errors.Wrapf(err, "service: s.model.AddProductCategory(ctx, categoryUUID=%q, productUUID=%q)", request.CategoryID, request.ProductID)
	}
	productCategory := ProductCategory{
		Object:     "product_category",
		ID:         row.UUID,
		ProductID:  row.ProductUUID,
		CategoryID: row.CategoryUUID,
		Pri:        row.Pri,
		Created:    row.Created,
		Modified:   row.Modified,
	}
	return &productCategory, nil
}

// GetProductCategory get a product to category association by id
func (s *Service) GetProductCategory(ctx context.Context, productCategoryID string) (*ProductCategory, error) {
	row, err := s.model.GetProductCategory(ctx, productCategoryID)
	if err != nil {
		if err == postgres.ErrProductCategoryNotFound {
			return nil, ErrProductCategoryNotFound
		}
		return nil, errors.Wrapf(err, "service: s.model.GetProductCategory(ctx, productCategoryUUID=%q", productCategoryID)
	}
	productCategory := ProductCategory{
		Object:     "product_category",
		ID:         row.UUID,
		ProductID:  row.ProductUUID,
		CategoryID: row.CategoryUUID,
		Pri:        row.Pri,
		Created:    row.Created,
		Modified:   row.Modified,
	}
	return &productCategory, nil
}

// DeleteProductCategory unlinks a product from a leaf category.
func (s *Service) DeleteProductCategory(ctx context.Context, productCategoryID string) error {
	err := s.model.DeleteProductCategory(ctx, productCategoryID)
	if err != nil {
		if err == postgres.ErrProductCategoryNotFound {
			return ErrProductCategoryNotFound
		}
		return err
	}
	return nil
}

// CreateProductCategoryRelations creates a set of catalog product
// associations either completing with all or failing with none
// being added.
func (s *Service) CreateProductCategoryRelations(ctx context.Context, cpas map[string][]string) error {
	err := s.model.CreateProductCategoryRelations(ctx, cpas)
	if err != nil {
		return errors.Wrap(err, "service: CreateProductCategoryRelationships")
	}
	return nil
}

// CreateProductCategoryAssoc associates an existing product to a catalog entry.
// func (s *Service) CreateProductCategoryAssoc(ctx context.Context, path, sku string) (*ProductCategoryAssoc, error) {
// 	cpa, err := s.model.CreateProductCategoryAssoc(ctx, path, sku)
// 	if err != nil {
// 		return nil, errors.Wrapf(err, "service: create product to category assoc sku=%q", sku)
// 	}
// 	scpa := ProductCategoryAssoc{
// 		Path:     cpa.Path,
// 		SKU:      cpa.SKU,
// 		Pri:      cpa.Pri,
// 		Created:  cpa.Created,
// 		Modified: cpa.Modified,
// 	}
// 	return &scpa, nil
// }

// HasProductCategoryRelations returns true if any product to category relations
// exist.
func (s *Service) HasProductCategoryRelations(ctx context.Context) (bool, error) {
	has, err := s.model.HasProductCategoryRelations(ctx)
	if err != nil {
		return false, errors.Wrap(err, "service: has product to category relations")
	}
	return has, nil
}

// An AssocProduct holds details of a product in the context of an Relationshipset.
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
		return nil, errors.Wrapf(err, "service: s.model.GetProductCategoryRelations(ctx) failed")
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
	createProductsCategories := make([]*postgres.CreateProductCategoryRow, 0, len(cpcs))
	for _, cp := range cpcs {
		c := postgres.CreateProductCategoryRow{
			ProductUUID:  cp.ProductID,
			CategoryUUID: cp.CategoryID,
		}
		createProductsCategories = append(createProductsCategories, &c)
	}

	list, err := s.model.UpdateProductsCategories(ctx, createProductsCategories)
	if err != nil {
		if err == postgres.ErrProductNotFound {
			return nil, ErrProductNotFound
		} else if err == postgres.ErrLeafCategoryNotFound {
			return nil, ErrLeafCategoryNotFound
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

// GetProductCategoryRelations returns all of the category product associations
// keyed by key. key has a value of either `id` or `path`
func (s *Service) GetProductCategoryRelations(ctx context.Context, key string) (map[string]*Assoc, error) {
	cpo, err := s.model.GetProductCategoryRelationsFull(ctx)
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

// UpdateProductCategoryRelations updates the category to product relations
// func (s *Service) UpdateProductCategoryRelations(ctx context.Context, cpo []*postgres.catalogProductAssoc) error {
// 	err := s.model.UpdateProductCategoryRelations(ctx, cpo)
// 	if err != nil {
// 		return err
// 	}
// 	return nil
// }

// DeleteAllProductCategoryRelations delete all catalog product associations.
func (s *Service) DeleteAllProductCategoryRelations(ctx context.Context) error {
	err := s.model.DeleteAllProductCategoryRelations(ctx)
	if err != nil {
		return errors.Wrapf(err, "service: delete product to category relations")
	}
	return nil
}
