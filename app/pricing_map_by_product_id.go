package app

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi"
	log "github.com/sirupsen/logrus"
)

// PricingMapByProductIDHandler creates a handler function that returns a
// map of tier refs to PriceEntry.
func (a *App) PricingMapByProductIDHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		contextLogger := log.WithContext(ctx)
		contextLogger.Info("App: PricingMapByProductIDHandler started")

		productID := chi.URLParam(r, "id")
		pmap, err := a.Service.PriceMap(ctx, productID)
		if err != nil {
			if err == sql.ErrNoRows {
				w.WriteHeader(http.StatusNotFound)
				return
			}
			contextLogger.Errorf("service PricingMapByProductID(ctx, productID=%q) error: %+v", productID, err)
			w.WriteHeader(http.StatusInternalServerError) // 500 Internal Server Error
			return
		}
		w.WriteHeader(http.StatusOK) // 200 OK
		json.NewEncoder(w).Encode(pmap)
	}
}
