package firebase

import (
	"context"

	"github.com/pkg/errors"
)

// ListAdmins returns a list of administrators
func (s *Service) ListAdmins(ctx context.Context) ([]*User, error) {
	admins, err := s.model.GetAllAdmins(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "GetAllAdmins(ctx) failed")
	}
	adms := make([]*User, 0, 8)
	for _, c := range admins {
		user := &User{
			ID:        c.UUID,
			UID:       c.UID,
			Role:      c.Role,
			Email:     c.Email,
			Firstname: c.Firstname,
			Lastname:  c.Lastname,
			Created:   c.Created,
			Modified:  c.Modified,
		}
		adms = append(adms, user)
	}
	return adms, nil
}
