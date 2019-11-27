package firebase

import (
	"context"
	"time"

	"bitbucket.org/andyfusniakteam/ecom-api-go/model/postgres"
	"github.com/pkg/errors"
)

// ErrProductSetNotFound error
var ErrProductSetNotFound = errors.New("service: product set not found")

// ProductSetItem represents a single item.
type ProductSetItem struct {
	Object       string    `json:"object"`
	ID           string    `json:"id"`
	ProductSetID string    `json:"product_set_id"`
	ProductID    string    `json:"product_id"`
	ProductPath  string    `json:"product_path"`
	ProductSKU   string    `json:"product_sku"`
	Created      time.Time `json:"created"`
	Modified     time.Time `json:"modified"`
}

// GetProductSetItems returns a list of product set items for a given product set id.
func (s *Service) GetProductSetItems(ctx context.Context, productSetID string) ([]*ProductSetItem, error) {
	rows, err := s.model.GetProductSetItems(ctx, productSetID)
	if err == postgres.ErrProductSetNotFound {
		return nil, ErrProductSetNotFound
	}
	if err != nil {
		return nil, errors.Wrapf(err, "s.model.GetProductSetItems(ctx, productSetID=%q)", productSetID)
	}

	productSetItems := make([]*ProductSetItem, 0, len(rows))
	for _, row := range rows {
		productSetItem := ProductSetItem{
			Object:       "product_set_item",
			ID:           row.UUID,
			ProductSetID: row.ProductSetUUID,
			ProductID:    row.ProductUUID,
			ProductPath:  row.ProductPath,
			ProductSKU:   row.ProductSKU,
			Created:      row.Created,
			Modified:     row.Modified,
		}
		productSetItems = append(productSetItems, &productSetItem)
	}
	return productSetItems, nil
}
