package firebase

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"net/http"
	"time"

	"bitbucket.org/andyfusniakteam/ecom-api-go/model/postgres"
	"cloud.google.com/go/pubsub"
	"github.com/btcsuite/btcutil/base58"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

// ErrWebhookExists webhook already exists
var ErrWebhookExists = errors.New("service: webhook exists")

// ErrWebhookNotFound webhook not found error
var ErrWebhookNotFound = errors.New("service: webhook not found")

// ErrEventTypeNotFound error
var ErrEventTypeNotFound = errors.New("service: event type not found")

// ErrWebhookPostFailed error
var ErrWebhookPostFailed = errors.New("service: http post failed")

const (
	// EventOrderCreated triggerred after an order has been placed.
	EventOrderCreated string = "order.created"
)

var eventTypes []string

var timeout = time.Duration(10 * time.Second)

var client *http.Client

func init() {
	eventTypes = []string{
		EventOrderCreated,
	}

	tr := &http.Transport{
		MaxIdleConnsPerHost: 10,
	}
	client = &http.Client{
		Transport: tr,
		Timeout:   timeout,
	}
}

// PubSubMsg holds the message section of a PubSubPayload.
type PubSubMsg struct {
	Attributes struct {
		Event     string `json:"event"`
		WebhookID string `json:"webhook_id,omitempty"`
	} `json:"attributes"`
	Data        string `json:"data"`
	MessageID   string `json:"message_id"`
	PublishTime string `json:"publish_time"`
}

// PubSubPayload represents the JSON payload from Cloud Pub/Sub
type PubSubPayload struct {
	Message      PubSubMsg `json:"message"`
	Subscription string    `json:"subscription"`
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

func contains(a []string, x string) bool {
	for _, n := range a {
		if x == n {
			return true
		}
	}
	return false
}

// BroadcastEvents publishes messages on Cloud Pub/Sub.
func (s *Service) BroadcastEvents(ctx context.Context, msg PubSubMsg) error {
	contextLogger := log.WithContext(ctx)
	contextLogger.Infof("service: BroadCastEvents(ctx, msg=%v) started", msg)

	webhooks, err := s.GetWebhooks(ctx)
	if err != nil {
		return errors.Wrap(err, "service: s.GetWebhooks(ctx) failed")
	}
	if len(webhooks) == 0 {
		log.Debug("service: no webhooks are registered - finished")
		return nil
	}
	log.Infof("service: %d webhooks are registered", len(webhooks))

	data, err := base64.StdEncoding.DecodeString(msg.Data)
	if err != nil {
		return errors.Wrapf(err, "service: base64.StdEncoding.DecodeString(s=%q)", msg.Data)
	}
	log.Debugf("service: base64 decoded string data=%q", data)

	// var v interface{}
	// err = json.Unmarshal(data, &v)
	// if err != nil {
	// 	return errors.Wrapf(err, "service: json.Unmarshal(data=%q, v=%v)", msg.Data, v)
	// }

	for _, wh := range webhooks {
		if !wh.Enabled {
			log.Infof("service: webhook not enabled - skipping id=%q, url=%s", wh.ID, wh.URL)
			continue
		}

		if contains(wh.Events, msg.Attributes.Event) {
			log.Debugf("service: webhook id=%q, url=%s matched on event %q", wh.ID, wh.URL, msg.Attributes.Event)
			publishResult := s.whBroadcastTopic.Publish(ctx, &pubsub.Message{
				Attributes: map[string]string{
					"event":      msg.Attributes.Event,
					"webhook_id": wh.ID,
				},
				Data: []byte(data),
			})
			serverID, err := publishResult.Get(ctx)
			if err != nil {
				log.Errorf("service: publishResult.Get(ctx) failed: %s", err.Error())
				return errors.Wrapf(err, "publishResult.Get(ctx) returned an error")
			}
			contextLogger.Infof("service: server-generated message ID %q", serverID)
			contextLogger.Infof("service: successfully published a message to pubsub")
		} else {
			log.Infof("service: webhook id=%q, url=%s, events=%q no matches on %q event", wh.ID, wh.URL, wh.Events, msg.Attributes.Event)
		}
	}
	return nil
}

// CallWebhook does a HTTP POST request to the webhook URL
// passing the event data.
func (s *Service) CallWebhook(ctx context.Context, webhookID, eventName string, rawData []byte) error {
	contextLogger := log.WithContext(ctx)
	contextLogger.Infof("service: CallWebhook(ctx, webhookID=%q, eventName=%q, data=%s) started", webhookID, eventName, string(rawData))

	webhook, err := s.model.GetWebhook(ctx, webhookID)
	if err == ErrWebhookNotFound {
		return ErrWebhookNotFound
	}
	if err != nil {
		return errors.Wrapf(err, "service: s.model.GetWebhook(ctx, webhookUUID=%q) failed", webhookID)
	}

	// Create a new HMAC by defining the hash type and the key (as byte array)
	h := hmac.New(sha256.New, []byte(webhook.SigningKey))
	_, err = h.Write(rawData)
	if err != nil {
		return errors.Wrapf(err, "service: h.Write(rawData=%q) failed", string(rawData))
	}
	sha := base64.StdEncoding.EncodeToString(h.Sum(nil))

	req, err := http.NewRequest("POST", webhook.URL, bytes.NewReader(rawData))
	if err != nil {
		return errors.Wrapf(err, "service: http.NewRequest(method=%q, url=%q, body=%s)", "POST", webhook.URL, string(rawData))
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Ecom-Hmac-SHA256", sha)

	res, err := client.Do(req)
	if err != nil {
		contextLogger.Warnf("service: client.Do(req=%+v) failed: %+v", req, err)
		return ErrWebhookPostFailed
	}
	defer res.Body.Close()

	// A push endpoint needs to handle incoming messages and return an HTTP
	// status code to indicate success or failure. A success response is
	// equivalent to acknowledging a messages. The status codes interpreted
	// as message acknowledgements by the ecom system are:
	// 200, 201, 202, 204, or 102.
	c := res.StatusCode
	if c == 200 || c == 201 || c == 202 || c == 204 || c == 102 {
		return nil
	}

	return ErrWebhookPostFailed
}
