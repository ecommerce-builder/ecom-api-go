package app

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi"
	log "github.com/sirupsen/logrus"
)

// PricingMapByTierHandler creates a handler function that returns a
// map of SKU to PricingEntry.
func (a *App) PricingMapByTierHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		contextLogger := log.WithContext(ctx)
		contextLogger.Info("App: PricingMapByTierHandler started")

		ref := chi.URLParam(r, "ref")
		pmap, err := a.Service.PricingMapByTier(ctx, ref)
		if err != nil {
			if err == sql.ErrNoRows {
				w.WriteHeader(http.StatusNotFound)
				return
			}
			contextLogger.Errorf("service PricingMapByTierHandler(ctx, %s) error: %+v", ref, err)
			w.WriteHeader(http.StatusInternalServerError) // 500 Internal Server Error
			return
		}
		w.WriteHeader(http.StatusOK) // 200 OK
		json.NewEncoder(w).Encode(pmap)
	}
}
