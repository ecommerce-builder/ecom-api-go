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
	Object    string    `json:"object"`
	ID        string    `json:"id"`
	ProductID string    `json:"product_id"`
	Path      string    `json:"path"`
	GSURL     string    `json:"gsurl"`
	Width     int       `json:"width"`
	Height    int       `json:"height"`
	Size      int       `json:"size"`
	Created   time.Time `json:"created"`
	Modified  time.Time `json:"modified"`
}

// CreateImageEntry creates a new image entry for a product with the given SKU.
func (s *Service) CreateImageEntry(ctx context.Context, productID string, path string) (*Image, error) {
	pc := postgres.CreateImage{
		W:     99999999,
		H:     99999999,
		Path:  path,
		GSURL: fmt.Sprintf("%s%s", "gs://", path),
		Typ:   "image/jpeg",
		Ori:   true,
		Pri:   10,
		Size:  0,
		Q:     100,
	}
	pi, err := s.model.CreateImageEntry(ctx, productID, &pc)
	if err != nil {
		if err == postgres.ErrProductNotFound {
			return nil, ErrProductNotFound
		}
		return nil, errors.Wrapf(err, "service: create image productID=%q, path=%q, entry failed", productID, path)
	}
	image := Image{
		Object:    "image",
		ID:        pi.UUID,
		ProductID: pi.ProductUUID,
		Path:      pi.Path,
		GSURL:     pi.GSURL,
		Width:     pi.W,
		Height:    pi.H,
		Size:      pi.Size,
		Created:   pi.Created,
		Modified:  pi.Modified,
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
	pi, err := s.model.GetProductImage(ctx, imageID)
	if err != nil {
		if err == postgres.ErrImageNotFound {
			return nil, ErrImageNotFound
		}
		return nil, errors.Wrapf(err, "service: GetProductImage(ctx, imageID=%q) failed", imageID)
	}
	image := Image{
		Object:    "image",
		ID:        pi.UUID,
		ProductID: pi.ProductUUID,
		Path:      pi.Path,
		GSURL:     pi.GSURL,
		Width:     pi.W,
		Height:    pi.H,
		Size:      pi.Size,
		Created:   pi.Created,
		Modified:  pi.Modified,
	}
	return &image, nil
}

// ListProductImages return a slice of Images.
func (s *Service) ListProductImages(ctx context.Context, productID string) ([]*Image, error) {
	pilist, err := s.model.GetImages(ctx, productID)
	if err != nil {
		if err == postgres.ErrProductNotFound {
			return nil, ErrProductNotFound
		}
		return nil, errors.Wrapf(err, "service: ListProductImages(ctx, productID=%q) failed", productID)
	}
	images := make([]*Image, 0, 8)
	for _, pi := range pilist {
		image := Image{
			Object:    "image",
			ID:        pi.UUID,
			ProductID: pi.ProductUUID,
			Path:      pi.Path,
			GSURL:     pi.GSURL,
			Width:     pi.W,
			Height:    pi.H,
			Size:      pi.Size,
			Created:   pi.Created,
			Modified:  pi.Modified,
		}
		images = append(images, &image)
	}
	return images, nil
}

// DeleteImage delete the image with the given ID.
func (s *Service) DeleteImage(ctx context.Context, imageID string) error {
	if err := s.model.DeleteProduct(ctx, imageID); err != nil {
		return errors.Wrapf(err, "service: DeleteProductImage(ctx, imageID=%q)", imageID)
	}
	return nil
}

// DeleteAllProductImages deletes all images associated to the product
// with the given SKU.
func (s *Service) DeleteAllProductImages(ctx context.Context, productID string) error {
	if err := s.model.DeleteAllProductImages(ctx, productID); err != nil {
		if err == postgres.ErrProductNotFound {
			return ErrProductNotFound
		}
		return errors.Wrapf(err, "service: DeleteAllProductImages(ctx, productID=%q)", productID)
	}
	return nil
}
