package app

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/go-chi/chi"
)

// PricingMapByTierHandler creates a handler function that returns a
// map of SKU to PricingEntry.
func (a *App) PricingMapByTierHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ref := chi.URLParam(r, "ref")
		pmap, err := a.Service.PricingMapByTier(r.Context(), ref)
		if err != nil {
			if err == sql.ErrNoRows {
				w.WriteHeader(http.StatusNotFound)
				return
			}
			fmt.Fprintf(os.Stderr, "service PricingMapByTierHandler(ctx, %s) error: %+v", ref, err)
			w.WriteHeader(http.StatusInternalServerError) // 500 Internal Server Error
			return
		}
		w.WriteHeader(http.StatusOK) // 200 OK
		json.NewEncoder(w).Encode(pmap)
	}
}
