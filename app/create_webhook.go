package app

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	service "bitbucket.org/andyfusniakteam/ecom-api-go/service/firebase"
	log "github.com/sirupsen/logrus"
)

type webhookEvents struct {
	Object string   `json:"object"`
	Data   []string `json:"data"`
}

type createWebhookRequestBody struct {
	URL    *string        `json:"url"`
	Events *webhookEvents `json:"events"`
}

func validateCreateWebhookRequest(request *createWebhookRequestBody) (bool, string) {
	// URL
	if request.URL == nil {
		return false, "attribute url must be set"
	}
	u, err := url.Parse(*request.URL)
	if err != nil {
		return false, fmt.Sprintf("error: %s", err)
	}
	if !u.IsAbs() {
		return false, "attribute url must be set to an absolute URL. Absolute means that it has a non-empty scheme."
	}
	if u.Scheme != "https" {
		return false, "attribute url must be set to a URL that uses the https:// protocol"
	}

	// Events
	if request.Events == nil {
		return false, "attribute events must be set to a list object"
	}

	return true, ""
}

// CreateWebhookHandler returns a function that creates a new webhook.
func (a *App) CreateWebhookHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		contextLogger := log.WithContext(ctx)
		contextLogger.Info("app: CreateWebhookHandler started")

		// parse the request body
		if r.Body == nil {
			// 400 Bad Request
			clientError(w, http.StatusBadRequest, ErrCodeBadRequest, "missing request body")
			return
		}
		request := createWebhookRequestBody{}
		dec := json.NewDecoder(r.Body)
		dec.DisallowUnknownFields()
		err := dec.Decode(&request)
		if err != nil {
			clientError(w, http.StatusBadRequest, ErrCodeBadRequest, err.Error())
			return
		}
		defer r.Body.Close()

		valid, message := validateCreateWebhookRequest(&request)
		if !valid {
			clientError(w, http.StatusBadRequest, ErrCodeBadRequest, message)
			return
		}

		// attempt to create the shipping tariff
		webhook, err := a.Service.CreateWebhook(ctx, *request.URL, request.Events.Data)
		if err == service.ErrWebhookExists {
			clientError(w, http.StatusConflict, ErrCodeWebhookExists, "webhook with this URL already exists")
			return
		}
		if err == service.ErrEventTypeNotFound {
			clientError(w, http.StatusNotFound, ErrCodeEventTypeNotFound, "one or more events were not recognised")
		}
		if err != nil {
			contextLogger.Errorf("app: a.Service.CreateWebhook(ctx, url=%q, events=%v) failed: %+v", *request.URL, request.Events.Data, err)
			w.WriteHeader(http.StatusInternalServerError) // 500
			return
		}

		w.WriteHeader(http.StatusCreated) // 201 Created
		json.NewEncoder(w).Encode(&webhook)
	}
}
