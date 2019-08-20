package app

import (
	"encoding/json"
	"net/http"

	service "bitbucket.org/andyfusniakteam/ecom-api-go/service/firebase"
	log "github.com/sirupsen/logrus"
)

// ListPricingTiersHandler creates a handler function that returns a
// list of pricing tiers.
func (a *App) ListPricingTiersHandler() http.HandlerFunc {
	type listPricingTiersResponse struct {
		Object string                 `json:"object"`
		Data   []*service.PricingTier `json:"data"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		contextLogger := log.WithContext(ctx)
		contextLogger.Info("App: ListPricingTiersHandler started")

		tiers, err := a.Service.GetPricingTiers(ctx)
		if err != nil {
			contextLogger.Errorf("service GetPricingTiers(ctx) error: %+v", err)
			w.WriteHeader(http.StatusInternalServerError) // 500 Internal Server Error
			return
		}

		list := listPricingTiersResponse{
			Object: "list",
			Data:   tiers,
		}
		w.WriteHeader(http.StatusOK) // 200 OK
		json.NewEncoder(w).Encode(&list)
	}
}
