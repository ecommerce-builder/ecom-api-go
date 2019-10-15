package app

import (
	"encoding/json"
	"net/http"

	service "bitbucket.org/andyfusniakteam/ecom-api-go/service/firebase"
	"github.com/go-chi/chi"
	log "github.com/sirupsen/logrus"
)

// GetShippingTariffHandler creates a handler function that returns a
// shipping tariff by id.
func (a *App) GetShippingTariffHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		contextLogger := log.WithContext(ctx)
		contextLogger.Info("app: GetShippingTariffHandler called")

		shippingTariffID := chi.URLParam(r, "id")
		if !IsValidUUID(shippingTariffID) {
			clientError(w, http.StatusBadRequest, ErrCodeBadRequest,
				"path parameters id must be a valid v4 uuid") // 400
			return
		}
		promoRule, err := a.Service.GetShippingTariff(ctx, shippingTariffID)
		if err == service.ErrShippingTariffNotFound {
			clientError(w, http.StatusNotFound, ErrCodeShippingTariffNotFound,
				"shopping tariff not found") // 404
			return
		}
		if err != nil {
			contextLogger.Errorf("app: a.Service.GetShippingTariff(ctx, shippingTariffID=%q) failed: %+v", shippingTariffID, err)
			w.WriteHeader(http.StatusInternalServerError) // 500
			return
		}
		w.WriteHeader(http.StatusOK) // 200
		json.NewEncoder(w).Encode(&promoRule)
	}
}
