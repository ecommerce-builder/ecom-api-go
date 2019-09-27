package firebase

import (
	"context"
	"time"

	"bitbucket.org/andyfusniakteam/ecom-api-go/model/postgres"
	"github.com/pkg/errors"
)

// ErrPPAssocGroupNotFound error.
var ErrPPAssocGroupNotFound = errors.New("service: product to product assoc group not found")

// ErrPPAssocGroupExists error
var ErrPPAssocGroupExists = errors.New("service: product to product assoc group exists")

// ErrPPAssocGroupContainsAssocs error occurs if attempting to delete a group than
// has existing associations in it.
var ErrPPAssocGroupContainsAssocs = errors.New("service: product to product assoc group contains associations")

// PPAssocGroup holds a product to product associations group.
type PPAssocGroup struct {
	Object   string    `json:"object"`
	ID       string    `json:"id"`
	Code     string    `json:"pp_assoc_group_code"`
	Name     string    `json:"name"`
	Created  time.Time `json:"created"`
	Modified time.Time `json:"modified"`
}

// CreateProductToProductAssocGroup creates a new product to product associations group.
func (s *Service) CreateProductToProductAssocGroup(ctx context.Context, code, name string) (*PPAssocGroup, error) {
	prow, err := s.model.AddPPAssocGroup(ctx, code, name)
	if err != nil {
		if err == postgres.ErrPPAssocGroupExists {
			return nil, ErrPPAssocGroupExists
		}
		return nil, errors.Wrapf(err, "service: s.model.AddPPAssocGroup(ctx, code=%q, name=%q) failed", code, name)
	}
	ppAssocGroup := PPAssocGroup{
		Object:   "pp_assoc_group",
		ID:       prow.UUID,
		Code:     prow.Code,
		Name:     prow.Name,
		Created:  prow.Created,
		Modified: prow.Modified,
	}
	return &ppAssocGroup, nil
}

// GetPPAssocGroup returns a single product to product association group.
func (s *Service) GetPPAssocGroup(ctx context.Context, ppAssocGroupID string) (*PPAssocGroup, error) {
	prow, err := s.model.GetPPAssocGroup(ctx, ppAssocGroupID)
	if err != nil {
		if err == postgres.ErrPPAssocGroupNotFound {
			return nil, ErrPPAssocGroupNotFound
		}
		return nil, errors.Wrapf(err, "service: s.model.GetPPAssocGroup(ctx, ppAssocGroupUUID=%q) failed", ppAssocGroupID)
	}
	ppAssocGroup := PPAssocGroup{
		Object:   "pp_assoc_group",
		ID:       prow.UUID,
		Code:     prow.Code,
		Name:     prow.Name,
		Created:  prow.Created,
		Modified: prow.Modified,
	}
	return &ppAssocGroup, nil
}

// GetPPAssocGroups returns a list of PPAssocGroups.
func (s *Service) GetPPAssocGroups(ctx context.Context) ([]*PPAssocGroup, error) {
	rows, err := s.model.GetPPAssocGroups(ctx)
	if err != nil {
		return nil, errors.Wrapf(err, "s.model.GetPriceLists(ctx) failed")
	}

	ppAssocGroups := make([]*PPAssocGroup, 0, len(rows))
	for _, row := range rows {
		g := PPAssocGroup{
			Object:   "pp_assoc_group",
			ID:       row.UUID,
			Code:     row.Code,
			Name:     row.Name,
			Created:  row.Created,
			Modified: row.Modified,
		}
		ppAssocGroups = append(ppAssocGroups, &g)
	}
	return ppAssocGroups, nil
}

// DeletePPAssocGroup deletes a single product to product associations group.
func (s *Service) DeletePPAssocGroup(ctx context.Context, ppAssocGroupID string) error {
	err := s.model.DeletePPAssocGroup(ctx, ppAssocGroupID)
	if err != nil {
		if err == postgres.ErrPPAssocGroupNotFound {
			return ErrPPAssocGroupNotFound
		} else if err == postgres.ErrPPAssocGroupContainsAssocs {
			return ErrPPAssocGroupContainsAssocs
		}
		return errors.Wrapf(err, "service: s.model.DeletePPAssocGroup(ctx, ppAssocGroupID=%q) failed", ppAssocGroupID)
	}
	return nil
}
