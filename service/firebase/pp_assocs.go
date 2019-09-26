package firebase

import (
	"context"
	"fmt"

	"bitbucket.org/andyfusniakteam/ecom-api-go/model/postgres"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

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

	fmt.Println(ppAssocsGroupID, productFrom, products)

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
