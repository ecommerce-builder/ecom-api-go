package firebase

import (
	"context"
	"time"

	"bitbucket.org/andyfusniakteam/ecom-api-go/model/postgres"
	"github.com/pkg/errors"
)

type ProductUpdate struct {
	EAN  string      `json:"ean" yaml:"ean"`
	Path string      `json:"path" yaml:"path"`
	Name string      `json:"name" yaml:"name"`
	Data ProductData `json:"data" yaml:"data"`
}

type ProductCreate struct {
	SKU  string      `json:"sku" yaml:"sku"`
	EAN  string      `json:"ean" yaml:"ean"`
	Path string      `json:"path" yaml:"path"`
	Name string      `json:"name" yaml:"name"`
	Data ProductData `json:"data" yaml:"data"`
}

type ProductData struct {
	Summary string `json:"summary" yaml:"summary"`
	Desc    string `json:"description" yaml:"description"`
	Spec    string `json:"specification" yaml:"specification"`
}

// Product contains all the fields that comprise a product in the catalog.
type Product struct {
	SKU      string                     `json:"sku" yaml:"sku,omitempty"`
	EAN      string                     `json:"ean" yaml:"ean"`
	Path     string                     `json:"path" yaml:"path"`
	Name     string                     `json:"name" yaml:"name"`
	Data     ProductData                `json:"data" yaml:"data"`
	Images   []*Image                   `json:"images" yaml:"images"`
	Pricing  map[string]*ProductPricing `json:"pricing" yaml:"pricing"`
	Created  time.Time                  `json:"created,omitempty"`
	Modified time.Time                  `json:"modified,omitempty"`
}

// CreateProduct create a new product if the product SKU does not already exist.
func (s *Service) CreateProduct(ctx context.Context, pc *ProductCreate) (*Product, error) {
	pu := &postgres.ProductUpdate{
		EAN:  pc.EAN,
		Path: pc.Path,
		Name: pc.Name,
		Data: postgres.ProductData{
			Summary: pc.Data.Summary,
			Desc:    pc.Data.Desc,
			Spec:    pc.Data.Spec,
		},
	}
	p, err := s.model.CreateProduct(ctx, pc.SKU, pu)
	if err != nil {
		return nil, errors.Wrapf(err, "create product %q failed", pc.SKU)
	}
	return &Product{
		SKU:  p.SKU,
		EAN:  p.EAN,
		Path: p.Path,
		Name: p.Name,
		Data: ProductData{
			Summary: p.Data.Summary,
			Desc:    p.Data.Desc,
			Spec:    p.Data.Spec,
		},
		Created:  p.Created,
		Modified: p.Modified,
	}, nil
}

// difference returns the elements in `a` that aren't in `b`.
func difference(a, b []string) []string {
	mb := map[string]bool{}
	for _, x := range b {
		mb[x] = true
	}
	var ab []string
	for _, x := range a {
		if _, ok := mb[x]; !ok {
			ab = append(ab, x)
		}
	}
	return ab
}

// ProductsExist accepts a slice of product SKUs and divides them into
// two lists of those that can exist in the system and those that are
// missing.
func (s *Service) ProductsExist(ctx context.Context, skus []string) (exists, missing []string, err error) {
	exists, err = s.model.ProductsExist(ctx, skus)
	if err != nil {
		return nil, nil, errors.Wrap(err, "service: ProductsExist")
	}
	missing = difference(skus, exists)
	return exists, missing, nil
}

// GetProduct gets a product given the SKU.
func (s *Service) GetProduct(ctx context.Context, sku string) (*Product, error) {
	p, err := s.model.GetProduct(ctx, sku)
	if err != nil {
		if err == postgres.ErrProductNotFound {
			return nil, err
		}
		return nil, errors.Wrapf(err, "model: GetProduct(ctx, %q) failed", sku)
	}
	images, err := s.ListProductImages(ctx, sku)
	if err != nil {
		return nil, errors.Wrapf(err, "service: ListProductImages(ctx, %q)", sku)
	}
	return &Product{
		SKU:  p.SKU,
		EAN:  p.EAN,
		Path: p.Path,
		Name: p.Name,
		Data: ProductData{
			Summary: p.Data.Summary,
			Desc:    p.Data.Desc,
			Spec:    p.Data.Spec,
		},
		Images:   images,
		Created:  p.Created,
		Modified: p.Modified,
	}, nil
}

// ListProducts returns a slice of all product SKUS.
func (s *Service) ListProducts(ctx context.Context) ([]string, error) {
	products, err := s.model.GetProducts(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "service: GetProduct")
	}
	skus := make([]string, 0, 256)
	for _, p := range products {
		skus = append(skus, p.SKU)
	}
	return skus, nil
}

func marshalProduct(a *Product, m *postgres.Product) {
	a.SKU = m.SKU
	a.EAN = m.EAN
	a.Path = m.Path
	a.Name = m.Name
	a.Data.Summary = m.Data.Summary
	a.Data.Desc = m.Data.Desc
	a.Data.Spec = m.Data.Spec
	a.Created = m.Created
	a.Modified = m.Modified
	return
}

// ProductExists return true if the given product exists.
func (s *Service) ProductExists(ctx context.Context, sku string) (bool, error) {
	exists, err := s.model.ProductExists(ctx, sku)
	if err != nil {
		return false, errors.Wrapf(err, "ProductExists(ctx, %q) failed", sku)
	}
	return exists, nil
}

// UpdateProduct updates a product by SKU.
func (s *Service) UpdateProduct(ctx context.Context, sku string, pu *ProductUpdate) (*Product, error) {
	update := &postgres.ProductUpdate{
		EAN:  pu.EAN,
		Path: pu.Path,
		Name: pu.Name,
		Data: postgres.ProductData{
			Summary: pu.Data.Summary,
			Desc:    pu.Data.Desc,
			Spec:    pu.Data.Spec,
		},
	}
	p, err := s.model.UpdateProduct(ctx, sku, update)
	if err != nil {
		return nil, errors.Wrapf(err, "update product sku=%q failed", sku)
	}
	ap := &Product{}
	marshalProduct(ap, p)
	return ap, nil
}

// DeleteProduct deletes the product with the given SKU.
func (s *Service) DeleteProduct(ctx context.Context, sku string) error {
	err := s.model.DeleteProduct(ctx, sku)
	if err != nil {
		return errors.Wrapf(err, "delete product sku=%q failed", sku)
	}
	return nil
}
