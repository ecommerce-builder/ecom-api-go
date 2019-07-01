package firebase

import (
	"context"

	"github.com/pkg/errors"
)

// ListAdmins returns a list of administrators
func (s *Service) ListAdmins(ctx context.Context) ([]*Customer, error) {
	admins, err := s.model.GetAllAdmins(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "GetAllAdmins(ctx) failed")
	}
	adms := make([]*Customer, 0, 8)
	for _, c := range admins {
		customer := &Customer{
			ID:        c.UUID,
			UID:       c.UID,
			Role:      c.Role,
			Email:     c.Email,
			Firstname: c.Firstname,
			Lastname:  c.Lastname,
			Created:   c.Created,
			Modified:  c.Modified,
		}
		adms = append(adms, customer)
	}
	return adms, nil
}

// DeleteAdmin deletes an administrator.
func (s *Service) DeleteAdmin(ctx context.Context, uuid string) error {
	authClient, err := s.fbApp.Auth(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to firebase auth client")
	}
	admin, err := s.model.GetAdmin(ctx, uuid)
	if err != nil {
		return errors.Wrapf(err, "get admin failed for uuid=%q", uuid)
	}
	if admin.Role != "admin" {
		return errors.Errorf("customer record with uuid=%q does not have role admin or could not be found", uuid)
	}
	if err = s.model.DeleteAdminByUUID(ctx, uuid); err != nil {
		return errors.Wrapf(err, "delete customer by uuid for uuid=%q failed", uuid)
	}
	if err = authClient.DeleteUser(ctx, admin.UID); err != nil {
		return errors.Wrapf(err, "firebase auth delete user failed for uid=%q", admin.UID)
	}
	return nil
}
