package app

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi"
	log "github.com/sirupsen/logrus"
)

// PricingMapBySKUHandler creates a handler function that returns a
// map of tier refs to PricingEntry.
func (a *App) PricingMapBySKUHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		contextLogger := log.WithContext(ctx)
		contextLogger.Info("App: PricingMapBySKUHandler started")

		sku := chi.URLParam(r, "sku")
		pmap, err := a.Service.PricingMapBySKU(ctx, sku)
		if err != nil {
			if err == sql.ErrNoRows {
				w.WriteHeader(http.StatusNotFound)
				return
			}
			contextLogger.Errorf("service PricingMapBySKU(ctx, %s) error: %+v", sku, err)
			w.WriteHeader(http.StatusInternalServerError) // 500 Internal Server Error
			return
		}
		w.WriteHeader(http.StatusOK) // 200 OK
		json.NewEncoder(w).Encode(pmap)
	}
}
