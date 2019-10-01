package app

import (
	"encoding/json"
	"net/http"

	service "bitbucket.org/andyfusniakteam/ecom-api-go/service/firebase"
	log "github.com/sirupsen/logrus"
)

// ListWebhooksHandler returns a http.HandlerFunc that returns
// a list of all webhooks.
func (a *App) ListWebhooksHandler() http.HandlerFunc {
	type response struct {
		Object string             `json:"object"`
		Data   []*service.Webhook `json:"data"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		contextLogger := log.WithContext(ctx)
		contextLogger.Info("app: ListWebhooksHandler called")

		webhooks, err := a.Service.GetWebhooks(ctx)
		if err != nil {
			contextLogger.Errorf("app: a.Service.GetWebhooks(ctx) failed: %+v", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		list := response{
			Object: "list",
			Data:   webhooks,
		}
		w.WriteHeader(http.StatusOK) // 200 OK
		json.NewEncoder(w).Encode(&list)
	}
}
