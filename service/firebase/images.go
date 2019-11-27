package firebase

import (
	"context"
	"fmt"
	"time"

	"bitbucket.org/andyfusniakteam/ecom-api-go/model/postgres"
	"github.com/pkg/errors"
)

// ErrImageNotFound is returned when any query
// for an image has no results in the resultset.
var ErrImageNotFound = errors.New("image not found")

// Image represents a product image.
type Image struct {
	Object      string    `json:"object"`
	ID          string    `json:"id"`
	ProductID   string    `json:"product_id"`
	ProductPath string    `json:"product_path"`
	ProductSKU  string    `json:"product_sku"`
	Path        string    `json:"path"`
	GSURL       string    `json:"gsurl"`
	Width       int       `json:"width"`
	Height      int       `json:"height"`
	Size        int       `json:"size"`
	Created     time.Time `json:"created"`
	Modified    time.Time `json:"modified"`
}

// CreateImage creates a new image for a product.
func (s *Service) CreateImage(ctx context.Context, productID string, path string) (*Image, error) {
	pc := postgres.CreateImage{
		ProductID: productID,
		W:         99999999,
		H:         99999999,
		Path:      path,
		GSURL:     fmt.Sprintf("%s%s", "gs://", path),
		Typ:       "image/jpeg",
		Ori:       true,
		Pri:       10,
		Size:      0,
		Q:         100,
	}
	i, err := s.model.CreateImage(ctx, &pc)
	if err == postgres.ErrProductNotFound {
		return nil, ErrProductNotFound
	}
	if err != nil {
		return nil, errors.Wrapf(err, "service: create image productID=%q, path=%q, entry failed", productID, path)
	}
	image := Image{
		Object:      "image",
		ID:          i.UUID,
		ProductID:   i.ProductUUID,
		ProductPath: i.ProductPath,
		ProductSKU:  i.ProductSKU,
		Path:        i.Path,
		GSURL:       i.GSURL,
		Width:       i.W,
		Height:      i.H,
		Size:        i.Size,
		Created:     i.Created,
		Modified:    i.Modified,
	}
	return &image, nil
}

// ImageUUIDExists returns true if the image with the given ID
// exists in the database. Note: it does not check if it exists
// in Google storage.
func (s *Service) ImageUUIDExists(ctx context.Context, id string) (bool, error) {
	exists, err := s.model.ImageUUIDExists(ctx, id)
	if err != nil {
		return false, errors.Wrapf(err, "service: ImageUUIDExists(ctx, %q) failed", id)
	}
	return exists, nil
}

// ImagePathExists returns true if the image with the given path
// exists in the database. Note: it does not check if it exists
// in Google storage.
func (s *Service) ImagePathExists(ctx context.Context, path string) (bool, error) {
	exists, err := s.model.ImagePathExists(ctx, path)
	if err != nil {
		return false, errors.Wrapf(err, "service: ImagePathExists(ctx, %q) failed", path)
	}
	return exists, nil
}

// GetImage returns an image by the given ID.
func (s *Service) GetImage(ctx context.Context, imageID string) (*Image, error) {
	i, err := s.model.GetProductImage(ctx, imageID)
	if err == postgres.ErrImageNotFound {
		return nil, ErrImageNotFound
	}
	if err != nil {
		return nil, errors.Wrapf(err, "service: GetProductImage(ctx, imageID=%q) failed", imageID)
	}
	image := Image{
		Object:      "image",
		ID:          i.UUID,
		ProductID:   i.ProductUUID,
		ProductPath: i.ProductPath,
		Path:        i.Path,
		GSURL:       i.GSURL,
		Width:       i.W,
		Height:      i.H,
		Size:        i.Size,
		Created:     i.Created,
		Modified:    i.Modified,
	}
	return &image, nil
}

// GetImagesByProductID return a slice of Images.
func (s *Service) GetImagesByProductID(ctx context.Context, productID string) ([]*Image, error) {
	pilist, err := s.model.GetImagesByProductUUID(ctx, productID)
	if err == postgres.ErrProductNotFound {
		return nil, ErrProductNotFound
	}
	if err != nil {
		return nil, errors.Wrapf(err, "service: s.model.GetImagesByProductUUID(ctx, productID=%q) failed", productID)
	}

	images := make([]*Image, 0, 8)
	for _, i := range pilist {
		image := Image{
			Object:      "image",
			ID:          i.UUID,
			ProductID:   i.ProductUUID,
			ProductPath: i.ProductPath,
			ProductSKU:  i.ProductSKU,
			Path:        i.Path,
			GSURL:       i.GSURL,
			Width:       i.W,
			Height:      i.H,
			Size:        i.Size,
			Created:     i.Created,
			Modified:    i.Modified,
		}
		images = append(images, &image)
	}
	return images, nil
}

// DeleteImage delete the image with the given ID.
func (s *Service) DeleteImage(ctx context.Context, imageID string) error {
	err := s.model.DeleteImage(ctx, imageID)
	if err == postgres.ErrImageNotFound {
		return ErrImageNotFound
	}
	if err != nil {
		return errors.Wrapf(err, "service: DeleteImage(ctx, imageID=%q)", imageID)
	}
	return nil
}

// DeleteAllProductImages deletes all images associated to the product
// with the given SKU.
func (s *Service) DeleteAllProductImages(ctx context.Context, productID string) error {
	err := s.model.DeleteAllProductImages(ctx, productID)
	if err == postgres.ErrProductNotFound {
		return ErrProductNotFound
	}
	if err != nil {
		return errors.Wrapf(err, "service: DeleteAllProductImages(ctx, productID=%q)", productID)
	}
	return nil
}
