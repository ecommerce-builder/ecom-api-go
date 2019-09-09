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
	ProductID string `json:"product_id"`
	Onhand    int    `json:"onhold"`
}

// Inventory holds inventory for a single product
type Inventory struct {
	Object      string    `json:"object"`
	ID          string    `json:"id"`
	ProductID   string    `json:"product_id"`
	ProductPath string    `json:"product_path"`
	ProductSKU  string    `json:"product_sku"`
	Onhand      int       `json:"onhold"`
	Created     time.Time `json:"created"`
	Modified    time.Time `json:"modified"`
}

// GetInventory returns a single Inventory by id.
func (s *Service) GetInventory(ctx context.Context, inventoryUUID string) (*Inventory, error) {
	contextLogger := log.WithContext(ctx)
	contextLogger.Info("service: GetInventory(ctx, inventoryUUID=%q) started", inventoryUUID)

	inv, err := s.model.GetInventoryByUUID(ctx, inventoryUUID)
	if err != nil {
		if err == postgres.ErrInventoryNotFound {
			return nil, ErrInventoryNotFound
		}
	}

	inventory := Inventory{
		Object:      "inventory",
		ID:          inv.UUID,
		ProductID:   inv.ProductUUID,
		ProductPath: inv.ProductPath,
		ProductSKU:  inv.ProductSKU,
		Onhand:      inv.Onhand,
		Created:     inv.Created,
		Modified:    inv.Modified,
	}

	return &inventory, nil
}

// GetAllInventory returns a slice of all inventory for all products.
func (s *Service) GetAllInventory(ctx context.Context) ([]*Inventory, error) {
	contextLogger := log.WithContext(ctx)
	contextLogger.Info("service: GetAllInventory(ctx) started")

	pinv, err := s.model.GetAllInventory(ctx)
	if err != nil {
		if err == postgres.ErrInventoryNotFound {
			return nil, ErrInventoryNotFound
		}
		return nil, errors.Wrapf(err, "s.model.GetAllInventory(ctx) failed")
	}

	inventory := make([]*Inventory, 0, len(pinv))
	for _, v := range pinv {
		inv := Inventory{
			Object:      "inventory",
			ID:          v.UUID,
			ProductID:   v.ProductUUID,
			ProductPath: v.ProductPath,
			ProductSKU:  v.ProductSKU,
			Onhand:      v.Onhand,
			Created:     v.Created,
			Modified:    v.Modified,
		}
		inventory = append(inventory, &inv)
	}
	return inventory, nil
}

// GetInventoryByProductID returns an Inventory for the given product.
func (s *Service) GetInventoryByProductID(ctx context.Context, productID string) (*Inventory, error) {
	contextLogger := log.WithContext(ctx)
	contextLogger.Infof("GetInventoryByProductID(ctx, productID=%q")

	v, err := s.model.GetInventoryByProductUUID(ctx, productID)
	if err != nil {
		if err == postgres.ErrProductNotFound {
			return nil, ErrProductNotFound
		}
	}

	inventory := Inventory{
		Object:      "inventory",
		ID:          v.UUID,
		ProductID:   v.ProductUUID,
		ProductPath: v.ProductPath,
		ProductSKU:  v.ProductSKU,
		Onhand:      v.Onhand,
		Created:     v.Created,
		Modified:    v.Modified,
	}
	return &inventory, nil
}

// UpdateInventory updates the inventory with the given inventoryID,
// to the new onhold value.
func (s *Service) UpdateInventory(ctx context.Context, inventoryID string, onhold int) (*Inventory, error) {
	contextLogger := log.WithContext(ctx)
	contextLogger.Infof("UpdateInventory(ctx, inventoryID=%q, onhold=%d)", inventoryID, onhold)

	v, err := s.model.UpdateInventoryByUUID(ctx, inventoryID, onhold)
	if err != nil {
		if err == postgres.ErrInventoryNotFound {
			return nil, ErrInventoryNotFound
		}
		return nil, errors.Wrapf(err, "s.model.UpdateInventoryByUUID(ctx, inventoryID, onhold) failed", inventoryID, onhold)
	}

	inventory := Inventory{
		Object:      "inventory",
		ID:          v.UUID,
		ProductID:   v.ProductUUID,
		ProductPath: v.ProductPath,
		ProductSKU:  v.ProductSKU,
		Onhand:      v.Onhand,
		Created:     v.Created,
		Modified:    v.Modified,
	}
	return &inventory, nil
}

// BatchUpdateInventory updates the inventory for multiple products in a single operations.
func (s *Service) BatchUpdateInventory(ctx context.Context, inventoryUpdates []*InventoryUpdateRequest) ([]*Inventory, error) {
	inventoryRows := make([]*postgres.InventoryRowUpdate, 0, len(inventoryUpdates))
	for _, i := range inventoryUpdates {
		pinv := postgres.InventoryRowUpdate{
			ProductUUID: i.ProductID,
			Onhand:      i.Onhand,
		}
		inventoryRows = append(inventoryRows, &pinv)
	}

	list, err := s.model.BatchUpdateInventory(ctx, inventoryRows)
	if err != nil {
		if err == postgres.ErrProductNotFound {
			return nil, ErrProductNotFound
		}
		return nil, errors.Wrap(err, "s.model.BatchUpdateInventory(ctx, inventoryRows) failed")
	}


	results := make([]*Inventory, 0, len(list))
	for _, i := range list {
		pc := Inventory{
			Object:       "inventory",
			ID:           i.UUID,
			ProductID:    i.ProductUUID,
			ProductPath:  i.ProductPath,
			ProductSKU:   i.ProductSKU,
			Onhand:       i.Onhand,
			Created:      i.Created,
			Modified:     i.Modified,
		}
		results = append(results, &pc)
	}
	return results, nil
}
