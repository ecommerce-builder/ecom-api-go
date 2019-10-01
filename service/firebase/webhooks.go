package firebase

import (
	"context"
	"crypto/rand"
	"time"

	"bitbucket.org/andyfusniakteam/ecom-api-go/model/postgres"
	"github.com/btcsuite/btcutil/base58"
	"github.com/pkg/errors"
)

// ErrWebhookExists webhook already exists
var ErrWebhookExists = errors.New("service: webhook exists")

// ErrWebhookNotFound webhook not found error
var ErrWebhookNotFound = errors.New("service: webhook not found")

// ErrEventTypeNotFound error
var ErrEventTypeNotFound = errors.New("service: event type not found")

const (
	// EventOrderCreated triggerred after an order has been placed.
	EventOrderCreated string = "order.created"
)

var eventTypes []string

func init() {
	eventTypes = []string{
		EventOrderCreated,
	}
}

// Webhook holds a single webhook.
type Webhook struct {
	Object     string    `json:"object"`
	ID         string    `json:"id"`
	SigningKey string    `json:"signing_key"`
	URL        string    `json:"url"`
	Events     []string  `json:"events"`
	Enabled    bool      `json:"enabled"`
	Created    time.Time `json:"created"`
	Modified   time.Time `json:"modified"`
}

// CreateWebhook creates a new webhook with the given url that triggers
// upon any of the events passed. events is a slice of event types
// who values must be a valid event type name. CreateWebhook returns
// the newly created Webhook or if an error occurs returns nil, along
// with an error. `ErrEventTypeNotFound` indicates an invalid event type
// name was passed. `ErrWebhookExists` is returned if the url is already
// used with another webhook.
func (s *Service) CreateWebhook(ctx context.Context, url string, events []string) (*Webhook, error) {
	// Check the given event name is a known event type
	eventTypeMap := make(map[string]bool)
	for _, v := range eventTypes {
		eventTypeMap[v] = true
	}
	for _, v := range events {
		if _, ok := eventTypeMap[v]; !ok {
			return nil, ErrEventTypeNotFound
		}
	}

	// Generate a cryptographically strong signing key
	// to be used later with HMAC SHA256.
	data := make([]byte, 32)
	_, err := rand.Read(data)
	if err != nil {
		return nil, errors.Wrap(err, "service: rand.Read(data)")
	}

	row, err := s.model.CreateWebhook(ctx, base58.Encode(data), url, events)
	if err != nil {
		if err == postgres.ErrWebhookExists {
			return nil, ErrWebhookExists
		}
		return nil, errors.Wrapf(err, "s.model.CreateWebhook(ctx, url=%q, events=%v) failed", url, events)
	}
	webhook := Webhook{
		Object:     "webhook",
		ID:         row.UUID,
		SigningKey: row.SigningKey,
		URL:        row.URL,
		Events:     row.Events,
		Enabled:    row.Enabled,
		Created:    row.Created,
		Modified:   row.Modified,
	}
	return &webhook, nil
}

// GetWebhook retrieves a single webhook by id. If the webhook id
// is not found GetWebhook returns `ErrWebhookNotFound`.
func (s *Service) GetWebhook(ctx context.Context, webhookID string) (*Webhook, error) {
	row, err := s.model.GetWebhook(ctx, webhookID)
	if err != nil {
		if err == postgres.ErrWebhookNotFound {
			return nil, ErrWebhookNotFound
		}
		return nil, errors.Wrapf(err, "s.model.GetWebhook(ctx, webhookUUID=%q) failed", webhookID)
	}
	webhook := Webhook{
		Object:     "webhook",
		ID:         row.UUID,
		SigningKey: row.SigningKey,
		URL:        row.URL,
		Events:     row.Events,
		Enabled:    row.Enabled,
		Created:    row.Created,
		Modified:   row.Modified,
	}
	return &webhook, nil
}

// GetWebhooks returns a slice of Webhooks for every webhook in the system.
func (s *Service) GetWebhooks(ctx context.Context) ([]*Webhook, error) {
	rows, err := s.model.GetWebhooks(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "s.model.GetWebhooks(ctx) failed")
	}

	webhooks := make([]*Webhook, 0, len(rows))
	for _, row := range rows {
		wh := Webhook{
			Object:     "webhook",
			ID:         row.UUID,
			SigningKey: row.SigningKey,
			URL:        row.URL,
			Events:     row.Events,
			Enabled:    row.Enabled,
			Created:    row.Created,
			Modified:   row.Modified,
		}
		webhooks = append(webhooks, &wh)
	}
	return webhooks, nil
}

// UpdateWebhook partially updates a webhook.
func (s *Service) UpdateWebhook(ctx context.Context, webhookUUID string, url *string, events []string, enabled *bool) (*Webhook, error) {
	// Check the given event name is a known event type
	eventTypeMap := make(map[string]bool)
	for _, v := range eventTypes {
		eventTypeMap[v] = true
	}
	for _, v := range events {
		if _, ok := eventTypeMap[v]; !ok {
			return nil, ErrEventTypeNotFound
		}
	}

	row, err := s.model.UpdateWebhook(ctx, webhookUUID, url, events, enabled)
	if err != nil {
		if err == postgres.ErrWebhookNotFound {
			return nil, ErrWebhookNotFound
		}
		return nil, errors.Wrapf(err, "service: s.model.UpdateWebhook(ctx, webhookUUID=%q, url=%v, events=%v, enabled=%v) failed", webhookUUID, url, events, enabled)
	}

	webhook := Webhook{
		Object:     "webhook",
		ID:         row.UUID,
		SigningKey: row.SigningKey,
		URL:        row.URL,
		Events:     row.Events,
		Enabled:    row.Enabled,
		Created:    row.Created,
		Modified:   row.Modified,
	}
	return &webhook, nil
}

// DeleteWebhook deletes a webhook by id.
func (s *Service) DeleteWebhook(ctx context.Context, webhookID string) error {
	err := s.model.DeleteWebhook(ctx, webhookID)
	if err != nil {
		if err == postgres.ErrWebhookNotFound {
			return ErrWebhookNotFound
		}
	}
	return nil
}
