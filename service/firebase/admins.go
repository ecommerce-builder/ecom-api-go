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

// DeleteAdmin deletes an administrator.
func (s *Service) DeleteAdmin(ctx context.Context, id string) error {
	authClient, err := s.fbApp.Auth(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to firebase auth client")
	}
	admin, err := s.model.GetAdmin(ctx, id)
	if err != nil {
		return errors.Wrapf(err, "get admin failed for id=%q", id)
	}
	if admin.Role != "admin" {
		return errors.Errorf("user record with id=%q does not have role admin or could not be found", id)
	}
	if err = s.model.DeleteAdminByUUID(ctx, id); err != nil {
		return errors.Wrapf(err, "delete user by id for id=%q failed", id)
	}
	if err = authClient.DeleteUser(ctx, admin.UID); err != nil {
		return errors.Wrapf(err, "firebase auth delete user failed for uid=%q", admin.UID)
	}
	return nil
}
