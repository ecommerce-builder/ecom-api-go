package app

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/go-chi/chi"
)

// GetTierPricingHandler creates a handler function that returns a
// product's pricing by SKU and tier ref.
func (a *App) GetTierPricingHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		sku := chi.URLParam(r, "sku")
		ref := chi.URLParam(r, "ref")
		pricing, err := a.Service.GetTierPricing(r.Context(), sku, ref)
		if err != nil {
			if err == sql.ErrNoRows {
				w.WriteHeader(http.StatusNotFound)
				return
			}
			fmt.Fprintf(os.Stderr, "service GetProductTierPricing(ctx, %s, %s) error: %+v", sku, ref, err)
			w.WriteHeader(http.StatusInternalServerError) // 500 Internal Server Error
			return
		}
		fmt.Println(pricing)
		w.WriteHeader(http.StatusOK) // 200 OK
		json.NewEncoder(w).Encode(*pricing)
	}
}
