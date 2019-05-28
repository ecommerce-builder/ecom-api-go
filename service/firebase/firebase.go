package firebase

import (
	"context"

	"bitbucket.org/andyfusniakteam/ecom-api-go/model/postgres"
	firebase "firebase.google.com/go"
	"firebase.google.com/go/auth"
)

// Service firebase implementation
type Service struct {
	model *postgres.PgModel
	fbApp *firebase.App
}

func NewService(model *postgres.PgModel, fbApp *firebase.App) *Service {
	return &Service{model, fbApp}
}

// Authenticate accepts a JSON Web Token, usually passed from the HTTP client and returns a auth.Token if valid or nil if
func (s *Service) Authenticate(ctx context.Context, jwt string) (*auth.Token, error) {
	authClient, err := s.fbApp.Auth(ctx)
	if err != nil {
		return nil, err
	}
	token, err := authClient.VerifyIDToken(ctx, jwt)
	if err != nil {
		return nil, err
	}
	return token, nil
}
