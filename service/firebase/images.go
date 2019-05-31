package firebase

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"bitbucket.org/andyfusniakteam/ecom-api-go/model/postgres"
	"github.com/pkg/errors"
)

// Image represents a product image.
type Image struct {
	UUID     string    `json:"uuid"`
	SKU      string    `json:"sku"`
	Path     string    `json:"path"`
	GSURL    string    `json:"gsurl"`
	Width    uint      `json:"width"`
	Height   uint      `json:"height"`
	Size     uint      `json:"size"`
	Created  time.Time `json:"created"`
	Modified time.Time `json:"modified"`
}

// CreateImageEntry creates a new image entry for a product with the given SKU.
func (s *Service) CreateImageEntry(ctx context.Context, sku, path string) (*Image, error) {
	pc := postgres.CreateImage{
		SKU:   sku,
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
	pi, err := s.model.CreateImageEntry(ctx, &pc)
	if err != nil {
		return nil, errors.Wrapf(err, "service: create image sku=%q, path=%q, entry failed", sku, path)
	}
	image := Image{
		UUID:     pi.UUID,
		SKU:      pi.SKU,
		Path:     pi.Path,
		GSURL:    pi.GSURL,
		Width:    pi.W,
		Height:   pi.H,
		Size:     pi.Size,
		Created:  pi.Created,
		Modified: pi.Modified,
	}
	return &image, nil
}

// ImageUUIDExists returns true if the image with the given UUID
// exists in the database. Note: it does not check if it exists
// in Google storage.
func (s *Service) ImageUUIDExists(ctx context.Context, uuid string) (bool, error) {
	exists, err := s.model.ImageUUIDExists(ctx, uuid)
	if err != nil {
		return false, errors.Wrapf(err, "service: ImageUUIDExists(ctx, %q) failed", uuid)
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

// GetImage returns an image by the given UUID.
func (s *Service) GetImage(ctx context.Context, uuid string) (*Image, error) {
	pi, err := s.model.GetProductImageByUUID(ctx, uuid)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, err
		}
		return nil, errors.Wrapf(err, "service: GetProductImageByUUID(ctx, %q) failed", uuid)
	}
	image := Image{
		UUID:     pi.UUID,
		SKU:      pi.SKU,
		Path:     pi.Path,
		GSURL:    pi.GSURL,
		Width:    pi.W,
		Height:   pi.H,
		Size:     pi.Size,
		Created:  pi.Created,
		Modified: pi.Modified,
	}
	return &image, nil
}

// ListProductImages return a slice of Images.
func (s *Service) ListProductImages(ctx context.Context, sku string) ([]*Image, error) {
	pilist, err := s.model.GetImagesBySKU(ctx, sku)
	if err != nil {
		return nil, errors.Wrapf(err, "service: ListProductImages(ctx, %q) failed", sku)
	}
	images := make([]*Image, 0, 8)
	for _, pi := range pilist {
		image := Image{
			UUID:     pi.UUID,
			SKU:      pi.SKU,
			Path:     pi.Path,
			GSURL:    pi.GSURL,
			Width:    pi.W,
			Height:   pi.H,
			Size:     pi.Size,
			Created:  pi.Created,
			Modified: pi.Modified,
		}
		images = append(images, &image)
	}
	return images, nil
}

// DeleteImage delete the image with the given UUID.
func (s *Service) DeleteImage(ctx context.Context, uuid string) error {
	if _, err := s.model.DeleteProductImage(ctx, uuid); err != nil {
		return err
	}
	return nil
}

// DeleteAllProductImages deletes all images associated to the product
// with the given SKU.
func (s *Service) DeleteAllProductImages(ctx context.Context, sku string) error {
	if _, err := s.model.DeleteAllProductImages(ctx, sku); err != nil {
		return errors.Wrapf(err, "service: DeleteAllProductImages(ctx, %q)", sku)
	}
	return nil
}
