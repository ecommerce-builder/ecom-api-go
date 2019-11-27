package app

import (
	"encoding/json"
	"net/http"

	service "bitbucket.org/andyfusniakteam/ecom-api-go/service/firebase"
	"github.com/go-chi/chi"
	log "github.com/sirupsen/logrus"
)

// GetOrderHandler returns a http.HandlerFunc that returns an order
// by object id.
func (a *App) GetOrderHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		contextLogger := log.WithContext(ctx)
		contextLogger.Info("app: GetOrderHandler called")

		orderID := chi.URLParam(r, "id")
		if !IsValidUUID(orderID) {
			contextLogger.Warn("app: 400 Bad Request - path parameter id must be a valid v4 uuid")
			clientError(w, http.StatusBadRequest, ErrCodeBadRequest,
				"path parameter id must be a valid v4 uuid") // 400
			return
		}

		order, err := a.Service.GetOrder(ctx, orderID)
		if err == service.ErrOrderNotFound {
			contextLogger.Warnf("app: 404 Not Found - order %s not found", orderID)
			clientError(w, http.StatusNotFound, ErrCodeOrderNotFound,
				"order not found") // 404
			return
		}
		if err == service.ErrOrderItemsNotFound {
			contextLogger.Warn("app: 404 Not Found - order items not found")
			clientError(w, http.StatusNotFound, ErrCodeOrderItemsNotFound,
				"order items not found") // 404
			return
		}
		if err != nil {
			contextLogger.Errorf("app: a.Service.GetOrder(ctx, orderID=%q): %+v",
				orderID, err)
			w.WriteHeader(http.StatusInternalServerError) // 500
			return
		}
		w.WriteHeader(http.StatusOK) // 200
		json.NewEncoder(w).Encode(&order)
	}
}
