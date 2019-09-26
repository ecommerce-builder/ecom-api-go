package firebase

import (
	"context"

	"bitbucket.org/andyfusniakteam/ecom-api-go/model/postgres"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

// ErrPPAssocNotFound error
var ErrPPAssocNotFound = errors.New("service: product to product association not found")

// ProductToUpdate a single product to update
type ProductToUpdate struct {
	ProductID string `json:"product_id"`
}

// BatchUpdatePPAssocs updates a set of product associations for a given product to product
// associations group and product_from combination.
func (s *Service) BatchUpdatePPAssocs(ctx context.Context, ppAssocsGroupID, productFrom string, productToSet []*ProductToUpdate) error {
	contextLogger := log.WithContext(ctx)
	contextLogger.Debugf("service: BatchUpdatePPAssocs(ctx context.Context, ppAssocsGroupID=%q, productFrom=%q, productToSet=%v)", ppAssocsGroupID, productFrom, productToSet)

	// build a slice of product ids
	var products []string
	for _, p := range productToSet {
		products = append(products, p.ProductID)
	}
	contextLogger.Debugf("postgres: product ids in the to set %v", products)

	if err := s.model.BatchUpdatePPAssocs(ctx, ppAssocsGroupID, productFrom, products); err != nil {
		if err == postgres.ErrPPAssocGroupNotFound {
			return ErrPPAssocGroupNotFound
		} else if err == postgres.ErrProductNotFound {
			return ErrProductNotFound
		}
		return errors.Wrapf(err, "service: s.model.BatchUpdatePPAssocs(ctx, ppAssocsGroupID=%q, productFrom=%q, products=%v)", ppAssocsGroupID, productFrom, products)
	}
	return nil
}

// DeletePPAssoc deletes a single product to product association
func (s *Service) DeletePPAssoc(ctx context.Context, ppAssocID string) error {
	contextLogger := log.WithContext(ctx)
	contextLogger.Debugf("service: DeletePPAssoc(ctx, ppAssocID=%q)", ppAssocID)

	if err := s.model.DeletePPAssoc(ctx, ppAssocID); err != nil {
		if err == postgres.ErrPPAssocNotFound {
			return ErrPPAssocNotFound
		}
		return errors.Wrapf(err, "s.model.DeletePPAssoc(ctx, ppAssocUUID=%q) failed", ppAssocID)
	}
	return nil
}
