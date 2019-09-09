package app

import (
	"encoding/json"
	"net/http"

	service "bitbucket.org/andyfusniakteam/ecom-api-go/service/firebase"
	"github.com/go-chi/chi"
	log "github.com/sirupsen/logrus"
)

// GetShippingTarrifHandler creates a handler function that returns a
// shipping tarrif by id.
func (a *App) GetShippingTarrifHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		contextLogger := log.WithContext(ctx)
		contextLogger.Info("app: GetShippingTarrifHandler called")

		shippingTarrifID := chi.URLParam(r, "id")
		promoRule, err := a.Service.GetShippingTarrif(ctx, shippingTarrifID)
		if err != nil {
			if err == service.ErrPromoRuleNotFound {
				clientError(w, http.StatusNotFound, ErrCodeShippingTarrifNotFound, "shopping tarrif not found")
				return
			}
			contextLogger.Errorf("app: a.Service.GetShippingTarrif(ctx, shippingTarrifID=%q) failed: %+v", shippingTarrifID, err)

			w.WriteHeader(http.StatusInternalServerError) // 500 Internal Server Error
		}
		w.WriteHeader(http.StatusOK) // 200 OK
		json.NewEncoder(w).Encode(&promoRule)
	}
}
