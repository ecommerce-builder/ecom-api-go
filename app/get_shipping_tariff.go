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
		promoRule, err := a.Service.GetShippingTariff(ctx, shippingTariffID)
		if err != nil {
			if err == service.ErrPromoRuleNotFound {
				clientError(w, http.StatusNotFound, ErrCodeShippingTariffNotFound, "shopping tariff not found")
				return
			}
			contextLogger.Errorf("app: a.Service.GetShippingTariff(ctx, shippingTariffID=%q) failed: %+v", shippingTariffID, err)

			w.WriteHeader(http.StatusInternalServerError) // 500 Internal Server Error
		}
		w.WriteHeader(http.StatusOK) // 200 OK
		json.NewEncoder(w).Encode(&promoRule)
	}
}
