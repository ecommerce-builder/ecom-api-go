package firebase

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"net/http"

	"cloud.google.com/go/pubsub"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

const (
	// EventServiceStarted event
	EventServiceStarted = "service.started"

	// EventUserCreated event
	EventUserCreated = "user.created"

	// EventAddressCreated event
	EventAddressCreated = "address.created"

	// EventAddressUpdated event
	EventAddressUpdated = "address.updated"

	// EventOrderCreated triggerred after an order has been placed.
	EventOrderCreated string = "order.created"

	// EventOrderUpdated event
	EventOrderUpdated string = "order.updated"
)

var validEvents map[string]struct{}

// {
//   "message_id": "136969346945",
//   "event": "order.created",
//   "webhook_id": "28a042cb-cadc-42f2-ae11-6add1d90968f",
//   "data": {
//   }
// }

// WebhookPayload holds structured data for the HTTP POST request
// on webhook callbacks.
type WebhookPayload struct {
	MessageID string      `json:"message_id"`
	Event     string      `json:"event"`
	WebhookID string      `json:"webhook_id"`
	Data      interface{} `json:"data"`
}

// ServiceStartedEventData simple message published upon
// startup of the HTTP Service.
type ServiceStartedEventData struct {
	Name string `json:"name"`
}

// DecodeEventData decodes event data.
func DecodeEventData(event string, data []byte) (interface{}, error) {
	var v interface{}
	if err := json.Unmarshal(data, &v); err != nil {
		return nil, errors.Wrapf(err, "service: json.Unmarshal(data=%v, &v) failed", data)
	}
	return v, nil
}

// CreateTopicAndSubscription sets up a topic and corresponding Push subscription.
// This function is used to intialise the system.
func CreateTopicAndSubscription(ctx context.Context, pubSubClient *pubsub.Client, topicName, subscrName, endpoint string) (*pubsub.Topic, error) {
	topic := pubSubClient.Topic(topicName)
	exists, err := topic.Exists(ctx)
	if err != nil {
		return nil, errors.Wrapf(err, "service: topic.Exists(ctx) failed")
	}
	if !exists {
		log.Printf("service: pubsub topic %q does not exist - creating it", topicName)
		topic, err = pubSubClient.CreateTopic(ctx, topicName)
		if err != nil {
			return nil, errors.Wrapf(err, "service: pubSubClient.CreateTopic(ctx, id=%q) failed", topicName)
		}
	}

	eventsSubscription := pubSubClient.Subscription(subscrName)
	exists, err = eventsSubscription.Exists(ctx)
	if err != nil {
		return nil, errors.Wrapf(err, "service: eventsSubscription.Exists(ctx) failed for subscrName=%q", subscrName)
	}
	if !exists {
		cfg := pubsub.SubscriptionConfig{
			Topic: topic,
			PushConfig: pubsub.PushConfig{
				Endpoint: endpoint,
			},
		}

		if _, err = pubSubClient.CreateSubscription(ctx, subscrName, cfg); err != nil {
			return nil, errors.Wrapf(err, "service: pubSubClient.CreateSubscription(ctx, subscrName=%q, cfg) failed", subscrName)
		}
		log.Infof("service: google pubsub subscription %q created", subscrName)
	} else {
		log.Infof("service: google pubsub subscription %q already exists", subscrName)
	}

	return topic, nil
}

// PublishTopicEvent publishes an event to Cloud Pub/Sub on the
// events topic.
func (s *Service) PublishTopicEvent(ctx context.Context, event string, v interface{}) error {
	contextLogger := log.WithContext(ctx)
	contextLogger.Infof("service: PublishTopicEvent(ctx, event=%q, data=%v) started", event, v)

	data, err := json.Marshal(v)
	if err != nil {
		return errors.Wrapf(err, "service: json.Marshal(v=%v) failed", v)
	}

	publishResult := s.eventsTopic.Publish(ctx, &pubsub.Message{
		Attributes: map[string]string{
			"event": event,
		},
		Data: data,
	})
	serverID, err := publishResult.Get(ctx)
	if err != nil {
		return errors.Wrap(err, "service: publishResult.Get(ctx) failed")
	}
	contextLogger.Infof("service: server-generated message ID %q", serverID)
	contextLogger.Infof("service: successfully published a message to pubsub events topic")
	return nil
}

// PublishBroadcastEvent publishes an event to Cloud Pub/Sub on the
// broadcast topic. Each event on the broadcast topic corresponds to
// an individual HTTP POST for a webhook URL endpoint setup by the
// caller.
func (s *Service) PublishBroadcastEvent(ctx context.Context, event, webhookID string, data []byte) error {
	contextLogger := log.WithContext(ctx)
	contextLogger.Infof("service: PublishBroadcastEvent(ctx, event=%q, webhookID=%q, data=%s) started", event, webhookID, data)

	publishResult := s.whBroadcastTopic.Publish(ctx, &pubsub.Message{
		Attributes: map[string]string{
			"event":      event,
			"webhook_id": webhookID,
		},
		Data: data,
	})
	serverID, err := publishResult.Get(ctx)
	if err != nil {
		log.Errorf("service: publishResult.Get(ctx) failed: %s", err.Error())
		return errors.Wrapf(err, "service: publishResult.Get(ctx) returned an error")
	}
	contextLogger.Infof("service: server-generated message ID %q", serverID)
	contextLogger.Infof("service: successfully published a message to pubsub broadcast topic")
	return nil
}

// CallWebhook does a HTTP POST request to the webhook URL
// passing the event data.
func (s *Service) CallWebhook(ctx context.Context, messageID, webhookID, event string, v interface{}) error {
	contextLogger := log.WithContext(ctx)
	contextLogger.Infof("service: CallWebhook(ctx, webhookID=%q, eventName=%q, v=%v) started", webhookID, event, v)

	webhook, err := s.model.GetWebhook(ctx, webhookID)
	if err == ErrWebhookNotFound {
		return ErrWebhookNotFound
	}
	if err != nil {
		return errors.Wrapf(err, "service: s.model.GetWebhook(ctx, webhookUUID=%q) failed", webhookID)
	}

	// build the payload and marshal the data attribute to JSON.
	payload := WebhookPayload{
		MessageID: messageID,
		Event:     event,
		WebhookID: webhookID,
		Data:      v,
	}
	data, err := json.Marshal(&payload)
	if err != nil {
		return errors.Wrapf(err, "service: json.Marshal(v=%v)", payload)
	}

	// Create a new HMAC by defining the hash type and the key (as byte array)
	h := hmac.New(sha256.New, []byte(webhook.SigningKey))
	_, err = h.Write(data)
	if err != nil {
		return errors.Wrapf(err, "service: h.Write(p=%s) failed", data)
	}
	sha := base64.StdEncoding.EncodeToString(h.Sum(nil))

	// Do the HTTP POST request passing the Hmac in the header
	req, err := http.NewRequest("POST", webhook.URL, bytes.NewReader(data))
	if err != nil {
		return errors.Wrapf(err, "service: http.NewRequest(method=%q, url=%q, body=%s)", "POST", webhook.URL, string(data))
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
