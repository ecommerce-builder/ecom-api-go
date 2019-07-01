package app

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi"
	log "github.com/sirupsen/logrus"
)

// GetTierPricingHandler creates a handler function that returns a
// product's pricing by SKU and tier ref.
func (a *App) GetTierPricingHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		contextLogger := log.WithContext(ctx)
		contextLogger.Info("App: GetTierPricingHandler called")

		sku := chi.URLParam(r, "sku")
		ref := chi.URLParam(r, "ref")
		pricing, err := a.Service.GetTierPricing(ctx, sku, ref)
		if err != nil {
			if err == sql.ErrNoRows {
				w.WriteHeader(http.StatusNotFound)
				return
			}
			contextLogger.Errorf("service GetProductTierPricing(ctx, %s, %s) error: %+v", sku, ref, err)
			w.WriteHeader(http.StatusInternalServerError) // 500 Internal Server Error
			return
		}
		w.WriteHeader(http.StatusOK) // 200 OK
		json.NewEncoder(w).Encode(*pricing)
	}
}
