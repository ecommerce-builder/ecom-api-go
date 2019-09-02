package firebase

import (
	"context"
	"time"

	"bitbucket.org/andyfusniakteam/ecom-api-go/model/postgres"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

// ErrProductNotFound is returned by GetProduct when the query
// for the product could not be found in the database.
var ErrProductNotFound = errors.New("service: product not found")

// ErrProductPathExists error
var ErrProductPathExists = errors.New("service: product path exists")

// ErrProductSKUExists error
var ErrProductSKUExists = errors.New("service: product sku exists")

// ProductImageRequestBody contains the product image data.
type ProductImageRequestBody struct {
	Path  string `json:"path"`
	Title string `json:"title"`
}

// ProductPricingRequestBody represents a product pricing entry to be added to updated.
type ProductPricingRequestBody struct {
	PriceListID string `json:"price_list_id"`
	UnitPrice   int    `json:"unit_price"`
}

// ProductCreateRequestBody contains fields required for creating a product.
type ProductCreateRequestBody struct {
	Path string `json:"path"`
	SKU  string `json:"sku"`
	Name string `json:"name"`
}

// ProductUpdateRequestBody contains fields required for updating a product.
type ProductUpdateRequestBody struct {
	Path string `json:"path"`
	SKU  string `json:"sku"`
	Name string `json:"name"`
}

type imageListContainer struct {
	Object string   `json:"object"`
	Data   []*Image `json:"data"`
}

type priceListContainer struct {
	Object string   `json:"object"`
	Data   []*Price `json:"data"`
}

// Product contains all the fields that comprise a product in the catalog.
type Product struct {
	Object   string              `json:"object"`
	ID       string              `json:"id"`
	Path     string              `json:"path"`
	SKU      string              `json:"sku"`
	Name     string              `json:"name"`
	Images   *imageListContainer `json:"images,omitempty"`
	Prices   *priceListContainer `json:"prices,omitempty"`
	Created  time.Time           `json:"created"`
	Modified time.Time           `json:"modified"`
}

// ProductList is a container for a list of product_slim objects.
type ProductList struct {
	Object string     `json:"object"`
	Data   []*Product `json:"data"`
}

// CreateProduct update and existing product by ID.
func (s *Service) CreateProduct(ctx context.Context, pc *ProductCreateRequestBody) (*Product, error) {
	// imagesReq := make([]*postgres.CreateImage, 0, 4)
	// for _, i := range pc.Images {
	// 	img := postgres.CreateImage{
	// 		W:     999999,
	// 		H:     999999,
	// 		Path:  i.Path,
	// 		Typ:   "image/jpeg",
	// 		Ori:   true,
	// 		Pri:   10,
	// 		Size:  0,
	// 		Q:     100,
	// 		GSURL: "gs://" + i.Path,
	// 	}
	// 	imagesReq = append(imagesReq, &img)
	// }
	// pricingReq := make([]*postgres.PriceEntry, 0, 4)
	// for _, r := range pc.Pricing {
	// 	item := postgres.PriceEntry{
	// 		PriceListUUID: string(r.PriceListID),
	// 		UnitPrice:     r.UnitPrice,
	// 	}
	// 	pricingReq = append(pricingReq, &item)
	// }
	p, err := s.model.CreateProduct(ctx, pc.Path, pc.SKU, pc.Name)
	if err != nil {
		if err == postgres.ErrPriceListNotFound {
			return nil, ErrPriceListNotFound
		} else if err == postgres.ErrProductPathExists {
			return nil, ErrProductPathExists
		} else if err == postgres.ErrProductSKUExists {
			return nil, ErrProductSKUExists
		}
		return nil, errors.Wrap(err, "CreateProduct(ctx) failed")
	}
	// images := make([]*Image, 0, 4)
	// for _, i := range p.Images {
	// 	img := Image{
	// 		Object:    "image",
	// 		ID:        i.UUID,
	// 		ProductID: i.ProductUUID,
	// 		Path:      i.Path,
	// 		GSURL:     i.GSURL,
	// 		Width:     i.W,
	// 		Height:    i.H,
	// 		Size:      i.Size,
	// 		Created:   i.Created,
	// 		Modified:  i.Modified,
	// 	}
	// 	images = append(images, &img)
	// }
	// prices := make(map[PriceListID]*Price)
	// for _, pr := range p.Prices {
	// 	price := Price{
	// 		UnitPrice: pr.UnitPrice,
	// 		Created:   pr.Created,
	// 		Modified:  pr.Modified,
	// 	}
	// 	prices[PriceListID(pr.UUID)] = &price
	// }
	return &Product{
		Object: "product",
		ID:     p.UUID,
		Path:   p.Path,
		SKU:    p.SKU,
		Name:   p.Name,
		Images: &imageListContainer{
			Object: "list",
			Data:   make([]*Image, 0),
		},
		Created:  p.Created,
		Modified: p.Modified,
	}, nil
}

// UpdateProduct updates an existing product by ID.
func (s *Service) UpdateProduct(ctx context.Context, productID string, pu *ProductUpdateRequestBody) (*Product, error) {
	// imagesReq := make([]*postgres.CreateImage, 0, 4)
	// for _, i := range pu.Images {
	// 	img := postgres.CreateImage{
	// 		W:     999999,
	// 		H:     999999,
	// 		Path:  i.Path,
	// 		Typ:   "image/jpeg",
	// 		Ori:   true,
	// 		Pri:   10,
	// 		Size:  0,
	// 		Q:     100,
	// 		GSURL: "gs://" + i.Path,
	// 	}
	// 	imagesReq = append(imagesReq, &img)
	// }
	// pricingReq := make([]*postgres.PriceEntry, 0, 4)
	// for _, r := range pu.Pricing {
	// 	item := postgres.PriceEntry{
	// 		PriceListUUID: string(r.PriceListID),
	// 		UnitPrice:     r.UnitPrice,
	// 	}
	// 	pricingReq = append(pricingReq, &item)
	// }
	update := &postgres.ProductUpdate{
		Path: pu.Path,
		SKU:  pu.SKU,
		Name: pu.Name,
	}
	p, err := s.model.UpdateProduct(ctx, productID, update)
	if err != nil {
		if err == postgres.ErrProductNotFound {
			return nil, ErrPriceListNotFound
		} else if err == postgres.ErrProductPathExists {
			return nil, ErrProductPathExists
		} else if err == postgres.ErrProductSKUExists {
			return nil, ErrProductSKUExists
		}
		return nil, errors.Wrapf(err, "UpdateProduct(ctx, productID=%v, ...) failed", productID)
	}
	// images := make([]*Image, 0, 4)
	// for _, i := range p.Images {
	// 	img := Image{
	// 		ID:        i.UUID,
	// 		ProductID: i.ProductUUID,
	// 		Path:      i.Path,
	// 		GSURL:     i.GSURL,
	// 		Width:     i.W,
	// 		Height:    i.H,
	// 		Size:      i.Size,
	// 		Created:   i.Created,
	// 		Modified:  i.Modified,
	// 	}
	// 	images = append(images, &img)
	// }
	// prices := make(map[PriceListID]*Price)
	// for _, pr := range p.Prices {
	// 	price := Price{
	// 		UnitPrice: pr.UnitPrice,
	// 		Created:   pr.Created,
	// 		Modified:  pr.Modified,
	// 	}
	// 	prices[PriceListID(pr.UUID)] = &price
	// }
	return &Product{
		Object:   "product",
		ID:       p.UUID,
		Path:     p.Path,
		SKU:      p.SKU,
		Name:     p.Name,
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

// ProductsExist accepts a slice of product uuids and divides them into
// two lists of those that can exist in the system and those that are
// missing.
func (s *Service) ProductsExist(ctx context.Context, productIDs []string) (exists, missing []string, err error) {
	exists, err = s.model.ProductsExist(ctx, productIDs)
	if err != nil {
		return nil, nil, errors.Wrap(err, "service: ProductsExist")
	}
	missing = difference(productIDs, exists)
	return exists, missing, nil
}

// GetProduct gets a product given the SKU.
func (s *Service) GetProduct(ctx context.Context, userID, productID string, includeImages, includePrices bool) (*Product, error) {
	contextLogger := log.WithContext(ctx)

	p, err := s.model.GetProduct(ctx, productID)
	if err != nil {
		if err == postgres.ErrProductNotFound {
			return nil, ErrProductNotFound
		}
		return nil, errors.Wrapf(err, "model: GetProduct(ctx, productID=%q) failed", productID)
	}

	// prices, err := s.PriceMap(ctx, p.UUID)
	// if err != nil {
	// 	return nil, errors.Wrapf(err, "service: PricingMapByProductID(ctx, productID=%q)", p.UUID)
	// }
	product := Product{
		Object:   "product",
		ID:       p.UUID,
		Path:     p.Path,
		SKU:      p.SKU,
		Name:     p.Name,
		Created:  p.Created,
		Modified: p.Modified,
	}

	// optional: include the images for this product
	if includeImages {
		contextLogger.Info("service: including the images for this product")
		images, err := s.GetImagesByProductID(ctx, p.UUID)
		if err != nil {
			return nil, errors.Wrapf(err, "service: ListProductImages(ctx, %q)", p.SKU)
		}

		product.Images = &imageListContainer{
			Object: "list",
			Data:   images,
		}
	}

	// optional: include the prices for this product
	if includePrices {
		contextLogger.Info("service: including the prices for this product")

		usrJoinRow, err := s.model.GetUserByUUID(ctx, userID)
		if err != nil {
			if err == postgres.ErrUserNotFound {
				return nil, ErrUserNotFound
			}
			return nil, errors.Wrapf(err, "service: s.model.GetUserByUUID(ctx, userUUID=%q) failed", userID)
		}

		priceListID := usrJoinRow.PriceListUUID
		prices, err := s.GetPrices(ctx, product.ID, priceListID)
		if err != nil {
			return nil, err
		}

		product.Prices = &priceListContainer{
			Object: "list",
			Data:   prices,
		}
	}
	return &product, nil
}

// ListProducts returns a slice of all product SKUS.
func (s *Service) ListProducts(ctx context.Context) ([]*Product, error) {
	products, err := s.model.GetProducts(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "service: GetProduct")
	}
	shortProducts := make([]*Product, 0, len(products))
	for _, p := range products {
		ps := Product{
			Object:   "product",
			ID:       p.UUID,
			Path:     p.Path,
			SKU:      p.SKU,
			Name:     p.Name,
			Created:  p.Created,
			Modified: p.Modified,
		}
		shortProducts = append(shortProducts, &ps)
	}
	return shortProducts, nil
}

// DeleteProduct deletes the product with the given UUID.
func (s *Service) DeleteProduct(ctx context.Context, uuid string) error {
	err := s.model.DeleteProduct(ctx, uuid)
	if err != nil {
		if err == postgres.ErrProductNotFound {
			return ErrProductNotFound
		}
		return errors.Wrapf(err, "delete product uuid=%q failed", uuid)
	}
	return nil
}
