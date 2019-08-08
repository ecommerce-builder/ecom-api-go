package firebase

import (
	"context"
	"time"

	"bitbucket.org/andyfusniakteam/ecom-api-go/model/postgres"
	"github.com/pkg/errors"
)

// ErrProductNotFound is returned by GetProduct when the query
// for the product could not be found in the database.
var ErrProductNotFound = errors.New("product not found")

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
	Object   string                    `json:"object"`
	ID       string                    `json:"id"`
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

// ProductSlim is a condensed representatio of a product
type ProductSlim struct {
	Object   string    `json:"object"`
	ID       string    `json:"id"`
	SKU      string    `json:"sku"`
	EAN      string    `json:"ean"`
	Path     string    `json:"path"`
	Name     string    `json:"name"`
	Created  time.Time `json:"created"`
	Modified time.Time `json:"modified"`
}

// UpdateProduct update and existing product by ID.
func (s *Service) UpdateProduct(ctx context.Context, productID string, pc *ProductCreate) (*Product, error) {
	imagesReq := make([]*postgres.CreateImage, 0, 4)
	for _, i := range pc.Images {
		img := postgres.CreateImage{
			ProductID: productID,
			W:         999999,
			H:         999999,
			Path:      i.Path,
			Typ:       "image/jpeg",
			Ori:       true,
			Pri:       10,
			Size:      0,
			Q:         100,
			GSURL:     "gs://" + i.Path,
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
	p, err := s.model.UpdateProduct(ctx, productID, update)
	if err != nil {
		return nil, errors.Wrapf(err, "UpdateProduct(ctx, productID=%q, ...) failed", productID)
	}
	images := make([]*Image, 0, 4)
	for _, i := range p.Images {
		img := Image{
			ID:       i.UUID,
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
		Object:  "product",
		ID:      p.UUID,
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
func (s *Service) GetProduct(ctx context.Context, productID string) (*Product, error) {
	p, err := s.model.GetProductByUUID(ctx, productID)
	if err != nil {
		if err == postgres.ErrProductNotFound {
			return nil, ErrProductNotFound
		}
		return nil, errors.Wrapf(err, "model: GetProduct(ctx, productID=%q) failed", productID)
	}
	images, err := s.ListProductImages(ctx, p.SKU)
	if err != nil {
		return nil, errors.Wrapf(err, "service: ListProductImages(ctx, %q)", p.SKU)
	}
	pricing, err := s.PricingMapBySKU(ctx, p.SKU)
	if err != nil {
		return nil, errors.Wrapf(err, "service: PricingMapBySKU(ctx, %q)", p.SKU)
	}
	return &Product{
		Object: "product",
		ID:     p.UUID,
		SKU:    p.SKU,
		EAN:    p.EAN,
		Path:   p.Path,
		Name:   p.Name,
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
func (s *Service) ListProducts(ctx context.Context) ([]*ProductSlim, error) {
	products, err := s.model.GetProducts(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "service: GetProduct")
	}
	shortProducts := make([]*ProductSlim, 0, len(products))
	for _, p := range products {
		ps := ProductSlim{
			Object:   "product_slim",
			ID:       p.UUID,
			SKU:      p.SKU,
			EAN:      p.EAN,
			Path:     p.Path,
			Name:     p.Name,
			Created:  p.Created,
			Modified: p.Modified,
		}
		shortProducts = append(shortProducts, &ps)
	}
	return shortProducts, nil
}

// ProductExists return true if the given product exists.
func (s *Service) ProductExists(ctx context.Context, uuid string) (bool, error) {
	exists, err := s.model.ProductExists(ctx, uuid)
	if err != nil {
		return false, errors.Wrapf(err, "ProductExists(ctx, uuid=%q) failed", uuid)
	}
	return exists, nil
}

// DeleteProduct deletes the product with the given UUID.
func (s *Service) DeleteProduct(ctx context.Context, uuid string) error {
	err := s.model.DeleteProduct(ctx, uuid)
	if err != nil {
		return errors.Wrapf(err, "delete product uuid=%q failed", uuid)
	}
	return nil
}
