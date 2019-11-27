package firebase

import (
	"context"
	"time"

	"bitbucket.org/andyfusniakteam/ecom-api-go/model/postgres"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

// ErrInventoryNotFound is returned when any query
// for inventory returns no results.
var ErrInventoryNotFound = errors.New("service: inventory not found")

// InventoryUpdateRequest for a single inventory update.
type InventoryUpdateRequest struct {
	ProductID   *string `json:"product_id"`
	Onhand      *int    `json:"onhand"`
	Overselling *bool   `json:"overselling"`
}

// Inventory holds inventory for a single product
type Inventory struct {
	Object      string    `json:"object"`
	ID          string    `json:"id"`
	ProductID   string    `json:"product_id"`
	ProductPath string    `json:"product_path"`
	ProductSKU  string    `json:"product_sku"`
	Onhand      int       `json:"onhand"`
	Overselling bool      `json:"overselling"`
	Created     time.Time `json:"created"`
	Modified    time.Time `json:"modified"`
}

// GetInventory returns a single Inventory by id.
func (s *Service) GetInventory(ctx context.Context, inventoryID string) (*Inventory, error) {
	contextLogger := log.WithContext(ctx)
	contextLogger.Infof("service: GetInventory(ctx, inventoryUUID=%q) started", inventoryID)

	row, err := s.model.GetInventoryByUUID(ctx, inventoryID)
	if err == postgres.ErrInventoryNotFound {
		return nil, ErrInventoryNotFound
	}
	if err != nil {
		return nil, errors.Wrapf(err, "service: s.model.GetInventoryByUUID(ctx, inventoryUUID=%q)", inventoryID)
	}

	inventory := Inventory{
		Object:      "inventory",
		ID:          row.UUID,
		ProductID:   row.ProductUUID,
		ProductPath: row.ProductPath,
		ProductSKU:  row.ProductSKU,
		Onhand:      row.Onhand,
		Overselling: row.Overselling,
		Created:     row.Created,
		Modified:    row.Modified,
	}
	return &inventory, nil
}

// GetAllInventory returns a slice of all inventory for all products.
func (s *Service) GetAllInventory(ctx context.Context) ([]*Inventory, error) {
	contextLogger := log.WithContext(ctx)
	contextLogger.Info("service: GetAllInventory(ctx) started")

	rows, err := s.model.GetAllInventory(ctx)
	if err == postgres.ErrInventoryNotFound {
		return nil, ErrInventoryNotFound
	}
	if err != nil {
		return nil, errors.Wrapf(err, "service: s.model.GetAllInventory(ctx) failed")
	}

	inventory := make([]*Inventory, 0, len(rows))
	for _, row := range rows {
		inv := Inventory{
			Object:      "inventory",
			ID:          row.UUID,
			ProductID:   row.ProductUUID,
			ProductPath: row.ProductPath,
			ProductSKU:  row.ProductSKU,
			Onhand:      row.Onhand,
			Overselling: row.Overselling,
			Created:     row.Created,
			Modified:    row.Modified,
		}
		inventory = append(inventory, &inv)
	}
	return inventory, nil
}

// GetInventoryByProductID returns an Inventory for the given product.
func (s *Service) GetInventoryByProductID(ctx context.Context, productID string) (*Inventory, error) {
	contextLogger := log.WithContext(ctx)
	contextLogger.Infof("GetInventoryByProductID(ctx, productID=%q) started", productID)

	row, err := s.model.GetInventoryByProductUUID(ctx, productID)
	if err == postgres.ErrProductNotFound {
		return nil, ErrProductNotFound
	}
	if err != nil {
		return nil, errors.Wrapf(err, "service: s.model.GetInventoryByProductUUID(ctx, productUUID=%q)", productID)
	}

	inventory := Inventory{
		Object:      "inventory",
		ID:          row.UUID,
		ProductID:   row.ProductUUID,
		ProductPath: row.ProductPath,
		ProductSKU:  row.ProductSKU,
		Onhand:      row.Onhand,
		Overselling: row.Overselling,
		Created:     row.Created,
		Modified:    row.Modified,
	}
	return &inventory, nil
}

// UpdateInventory updates the inventory with the given inventoryID,
// to the new onhand value.
func (s *Service) UpdateInventory(ctx context.Context, inventoryID string, onhand *int, overselling *bool) (*Inventory, error) {
	contextLogger := log.WithContext(ctx)
	contextLogger.Infof("service: UpdateInventory(ctx, inventoryID=%q, onhand=%v, overselling=%v) started", inventoryID, onhand, overselling)

	row, err := s.model.UpdateInventoryByUUID(ctx, inventoryID, onhand, overselling)
	if err == postgres.ErrInventoryNotFound {
		return nil, ErrInventoryNotFound
	}
	if err != nil {
		return nil, errors.Wrapf(err, "service: s.model.UpdateInventoryByUUID(ctx, inventoryID=%q, onhand=%d) failed", inventoryID, onhand)
	}

	inventory := Inventory{
		Object:      "inventory",
		ID:          row.UUID,
		ProductID:   row.ProductUUID,
		ProductPath: row.ProductPath,
		ProductSKU:  row.ProductSKU,
		Onhand:      row.Onhand,
		Overselling: row.Overselling,
		Created:     row.Created,
		Modified:    row.Modified,
	}
	return &inventory, nil
}

// BatchUpdateInventory updates the inventory for multiple products in a single operations.
func (s *Service) BatchUpdateInventory(ctx context.Context, inventoryUpdates []*InventoryUpdateRequest) ([]*Inventory, error) {
	contextLogger := log.WithContext(ctx)
	contextLogger.Info("service: BatchUpdateInventory(ctx, inventoryUpdates) started")

	inventoryRows := make([]*postgres.InventoryRowUpdate, 0, len(inventoryUpdates))
	for _, i := range inventoryUpdates {
		pinv := postgres.InventoryRowUpdate{
			ProductUUID: *i.ProductID,
			Onhand:      *i.Onhand,
			Overselling: *i.Overselling,
		}
		inventoryRows = append(inventoryRows, &pinv)
	}

	rows, err := s.model.BatchUpdateInventory(ctx, inventoryRows)
	if err == postgres.ErrProductNotFound {
		return nil, ErrProductNotFound
	}
	if err != nil {
		return nil, errors.Wrap(err, "service: s.model.BatchUpdateInventory(ctx, inventoryRows) failed")
	}

	inventory := make([]*Inventory, 0, len(rows))
	for _, row := range rows {
		pc := Inventory{
			Object:      "inventory",
			ID:          row.UUID,
			ProductID:   row.ProductUUID,
			ProductPath: row.ProductPath,
			ProductSKU:  row.ProductSKU,
			Onhand:      row.Onhand,
			Overselling: row.Overselling,
			Created:     row.Created,
			Modified:    row.Modified,
		}
		inventory = append(inventory, &pc)
	}
	return inventory, nil
}
