package firebase

import (
	"context"

	"bitbucket.org/andyfusniakteam/ecom-api-go/model/postgres"
	"cloud.google.com/go/pubsub"
	firebase "firebase.google.com/go"
	"firebase.google.com/go/auth"
	"github.com/pkg/errors"
)

// Service firebase implementation
type Service struct {
	model            *postgres.PgModel
	fbApp            *firebase.App
	eventsTopic      *pubsub.Topic
	whBroadcastTopic *pubsub.Topic
}

// NewService creates a new Service
func NewService(model *postgres.PgModel, fbApp *firebase.App, eventsTopic, whBroadcastTopic *pubsub.Topic) *Service {
	return &Service{
		model:            model,
		fbApp:            fbApp,
		eventsTopic:      eventsTopic,
		whBroadcastTopic: whBroadcastTopic,
	}
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

// GetSchemaVersion returns the underlying database schema version as string.
func (s *Service) GetSchemaVersion(ctx context.Context) (*string, error) {
	version, err := s.model.GetSchemaVersion(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "service: GetSchemaVersion() failed")
	}
	return version, nil
}
