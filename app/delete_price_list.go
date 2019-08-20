package app

import (
	"net/http"

	"github.com/go-chi/chi"
	log "github.com/sirupsen/logrus"
)

// DeleteTierPricingHandler creates a handler function that deletes
// a tier pricing by SKU and tier ref.
func (a *App) DeleteTierPricingHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		contextLogger := log.WithContext(ctx)
		contextLogger.Info("App: DeleteTierPricingHandler started")

		sku := chi.URLParam(r, "sku")
		ref := chi.URLParam(r, "ref")
		if err := a.Service.DeleteTierPricing(ctx, sku, ref); err != nil {
			contextLogger.Errorf("service DeleteTierPricingHandler(ctx, %s, %s) error: %+v", sku, ref, err)
			w.WriteHeader(http.StatusInternalServerError) // 500 Internal Server Error
			return
		}
		w.Header().Del("Content-Type")
		w.WriteHeader(http.StatusNoContent) // 204 No Content
	}
}
