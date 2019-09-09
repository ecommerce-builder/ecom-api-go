package app

import (
	"encoding/json"
	"net/http"

	service "bitbucket.org/andyfusniakteam/ecom-api-go/service/firebase"
	log "github.com/sirupsen/logrus"
)

// ListShippingTarrifsHandler creates a handler function that returns a
// list of shipping tarrifs.
func (a *App) ListShippingTarrifsHandler() http.HandlerFunc {
	type listPromoRulesResponse struct {
		Object string                    `json:"object"`
		Data   []*service.ShippingTarrif `json:"data"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		contextLogger := log.WithContext(ctx)
		contextLogger.Info("app: ListShippingTarrifsHandler started")

		shippingTarrifs, err := a.Service.GetShippingTarrifs(ctx)
		if err != nil {
			contextLogger.Errorf("app: a.Service.GetShippingTarrifs(ctx) error: %+v", err)
			w.WriteHeader(http.StatusInternalServerError) // 500 Internal Server Error
			return
		}

		list := listPromoRulesResponse{
			Object: "list",
			Data:   shippingTarrifs,
		}
		w.WriteHeader(http.StatusOK) // 200 OK
		json.NewEncoder(w).Encode(&list)
	}
}
