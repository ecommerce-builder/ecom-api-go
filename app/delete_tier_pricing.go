package app

import (
	"fmt"
	"net/http"
	"os"

	"github.com/go-chi/chi"
)

// DeleteTierPricingHandler creates a handler function that deletes
// a tier pricing by SKU and tier ref.
func (a *App) DeleteTierPricingHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		sku := chi.URLParam(r, "sku")
		ref := chi.URLParam(r, "ref")
		if err := a.Service.DeleteTierPricing(r.Context(), sku, ref); err != nil {
			fmt.Fprintf(os.Stderr, "service DeleteTierPricingHandler(ctx, %s, %s) error: %+v", sku, ref, err)
			w.WriteHeader(http.StatusInternalServerError) // 500 Internal Server Error
			return
		}
		w.Header().Del("Content-Type")
		w.WriteHeader(http.StatusNoContent) // 204 No Content
	}
}
