package app

import (
	"encoding/json"
	"net/http"

	service "bitbucket.org/andyfusniakteam/ecom-api-go/service/firebase"
	log "github.com/sirupsen/logrus"
)

// ListOrdersHandler creates a handler function that returns a
// list of orders.
func (a *App) ListOrdersHandler() http.HandlerFunc {
	type response struct {
		Object string           `json:"object"`
		Data   []*service.Order `json:"data"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		contextLogger := log.WithContext(ctx)
		contextLogger.Info("app: ListOrdersHandler called")

		orders, err := a.Service.GetOrders(ctx)
		if err != nil {
			contextLogger.Errorf("app: a.Service.GetOrders(ctx) failed: %+v", err)
			w.WriteHeader(http.StatusInternalServerError) // 500
			return
		}
		contextLogger.Debugf("app: a.Service.GetOrders(ctx) return %d orders",
			len(orders))

		list := response{
			Object: "list",
			Data:   orders,
		}
		w.WriteHeader(http.StatusOK) // 200
		json.NewEncoder(w).Encode(&list)
	}
}
