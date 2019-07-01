package firebase

import (
	"context"
	"time"

	"bitbucket.org/andyfusniakteam/ecom-api-go/model/postgres"
	"github.com/pkg/errors"
)

// ImageEntry contains the product image data.
type ImageEntry struct {
	Path  string `json:"path"`
	Title string `json:"title"`
}

// ProductContent contains the variable JSON data of the product
type ProductContent struct {
	Meta struct {
		Title       string `json:"title"`
		Description string `json:"description"`
	} `json:"meta"`
	Videos        []string `json:"videos"`
	Manuals       []string `json:"manuals"`
	Software      []string `json:"software"`
	Description   string   `json:"description"`
	Specification string   `json:"specification"`
	InTheBox      string   `json:"in_the_box"`
}

// ProductCreate contains fields required for creating a product.
type ProductCreate struct {
	EAN     string            `json:"ean"`
	Path    string            `json:"path"`
	Name    string            `json:"name"`
	Images  []*ImageEntry     `json:"images"`
	Pricing []*ProductPricing `json:"pricing"`
	Content ProductContent    `json:"content"`
}

// Product contains all the fields that comprise a product in the catalog.
type Product struct {
	SKU      string                    `json:"sku"`
	EAN      string                    `json:"ean"`
	Path     string                    `json:"path"`
	Name     string                    `json:"name"`
	Images   []*Image                  `json:"images"`
	Pricing  map[TierRef]*PricingEntry `json:"pricing"`
	Content  ProductContent            `json:"content"`
	Created  time.Time                 `json:"created,omitempty"`
	Modified time.Time                 `json:"modified,omitempty"`
}

// ReplaceProduct create a new product if the product SKU does not
// already exist, or updates it if it does.
func (s *Service) ReplaceProduct(ctx context.Context, sku string, pc *ProductCreate) (*Product, error) {
	imagesReq := make([]*postgres.CreateImage, 0, 4)
	for _, i := range pc.Images {
		img := postgres.CreateImage{
			SKU:   sku,
			W:     999999,
			H:     999999,
			Path:  i.Path,
			Typ:   "image/jpeg",
			Ori:   true,
			Pri:   10,
			Size:  0,
			Q:     100,
			GSURL: "gs://" + i.Path,
		}
		imagesReq = append(imagesReq, &img)
	}
	pricingReq := make([]*postgres.ProductPricingEntry, 0, 4)
	for _, r := range pc.Pricing {
		item := postgres.ProductPricingEntry{
			TierRef:   string(r.TierRef),
			UnitPrice: r.UnitPrice,
		}
		pricingReq = append(pricingReq, &item)
	}
	update := &postgres.ProductCreateUpdate{
		EAN:     pc.EAN,
		Path:    pc.Path,
		Name:    pc.Name,
		Images:  imagesReq,
		Pricing: pricingReq,
		Content: postgres.ProductContent{
			Meta:          pc.Content.Meta,
			Videos:        pc.Content.Videos,
			Manuals:       pc.Content.Manuals,
			Software:      pc.Content.Software,
			Description:   pc.Content.Description,
			Specification: pc.Content.Specification,
			InTheBox:      pc.Content.InTheBox,
		},
	}
	p, err := s.model.UpdateProduct(ctx, sku, update)
	if err != nil {
		return nil, errors.Wrapf(err, "UpdateProduct(ctx, sku=%q, ...) failed", sku)
	}
	images := make([]*Image, 0, 4)
	for _, i := range p.Images {
		img := Image{
			ID:       i.UUID,
			SKU:      i.SKU,
			Path:     i.Path,
			GSURL:    i.GSURL,
			Width:    i.W,
			Height:   i.H,
			Size:     i.Size,
			Created:  i.Created,
			Modified: i.Modified,
		}
		images = append(images, &img)
	}
	pricing := make(map[TierRef]*PricingEntry)
	for _, pr := range p.Pricing {
		price := PricingEntry{
			UnitPrice: pr.UnitPrice,
			Created:   pr.Created,
			Modified:  pr.Modified,
		}
		pricing[TierRef(pr.TierRef)] = &price
	}
	return &Product{
		SKU:     p.SKU,
		EAN:     p.EAN,
		Path:    p.Path,
		Name:    p.Name,
		Images:  images,
		Pricing: pricing,
		Content: ProductContent{
			Meta:          pc.Content.Meta,
			Videos:        pc.Content.Videos,
			Manuals:       pc.Content.Manuals,
			Software:      pc.Content.Software,
			Description:   pc.Content.Description,
			Specification: pc.Content.Specification,
			InTheBox:      pc.Content.InTheBox,
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
	pricing, err := s.PricingMapBySKU(ctx, sku)
	if err != nil {
		return nil, errors.Wrapf(err, "service: PricingMapBySKU(ctx, %q)", sku)
	}
	return &Product{
		SKU:  p.SKU,
		EAN:  p.EAN,
		Path: p.Path,
		Name: p.Name,
		Content: ProductContent{
			Description:   p.Content.Description,
			Specification: p.Content.Specification,
		},
		Images:   images,
		Pricing:  pricing,
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

// ProductExists return true if the given product exists.
func (s *Service) ProductExists(ctx context.Context, sku string) (bool, error) {
	exists, err := s.model.ProductExists(ctx, sku)
	if err != nil {
		return false, errors.Wrapf(err, "ProductExists(ctx, %q) failed", sku)
	}
	return exists, nil
}

// DeleteProduct deletes the product with the given SKU.
func (s *Service) DeleteProduct(ctx context.Context, sku string) error {
	err := s.model.DeleteProduct(ctx, sku)
	if err != nil {
		return errors.Wrapf(err, "delete product sku=%q failed", sku)
	}
	return nil
}
