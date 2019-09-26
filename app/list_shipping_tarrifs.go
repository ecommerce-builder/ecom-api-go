package app

import (
	"encoding/json"
	"net/http"

	service "bitbucket.org/andyfusniakteam/ecom-api-go/service/firebase"
	log "github.com/sirupsen/logrus"
)

// ListShippingTariffsHandler creates a handler function that returns a
// list of shipping tariffs.
func (a *App) ListShippingTariffsHandler() http.HandlerFunc {
	type listPromoRulesResponse struct {
		Object string                    `json:"object"`
		Data   []*service.ShippingTariff `json:"data"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		contextLogger := log.WithContext(ctx)
		contextLogger.Info("app: ListShippingTariffsHandler started")

		shippingTariffs, err := a.Service.GetShippingTariffs(ctx)
		if err != nil {
			contextLogger.Errorf("app: a.Service.GetShippingTariffs(ctx) error: %+v", err)
			w.WriteHeader(http.StatusInternalServerError) // 500 Internal Server Error
			return
		}

		list := listPromoRulesResponse{
			Object: "list",
			Data:   shippingTariffs,
		}
		w.WriteHeader(http.StatusOK) // 200 OK
		json.NewEncoder(w).Encode(&list)
	}
}
