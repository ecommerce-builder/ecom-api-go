package app

import (
	"encoding/json"
	"net/http"

	log "github.com/sirupsen/logrus"
)

// GetTierPricingHandler creates a handler function that returns a
// product's pricing by SKU and tier ref.
func (a *App) GetTierPricingHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		contextLogger := log.WithContext(ctx)
		contextLogger.Info("App: GetTierPricingHandler called")

		// productID := chi.URLParam(r, "id")
		// ref := ""
		// price, err := a.Service.GetProductPrice(ctx, productID, ref)
		// if err != nil {
		// 	if err == sql.ErrNoRows {
		// 		w.WriteHeader(http.StatusNotFound)
		// 		return
		// 	}
		// 	contextLogger.Errorf("service GetProductPrice(ctx, %q, %q) error: %+v", productID, ref, err)
		// 	w.WriteHeader(http.StatusInternalServerError) // 500 Internal Server Error
		// 	return
		// }
		w.WriteHeader(http.StatusOK) // 200 OK
		json.NewEncoder(w).Encode(nil)
	}
}
