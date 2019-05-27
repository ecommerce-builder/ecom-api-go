package app

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/go-chi/chi"
)

// ListPricingByTierHandler creates a handler function that returns a
// map of SKU to pricing entries.
func (a *App) ListPricingByTierHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ref := chi.URLParam(r, "ref")
		pmap, err := a.Service.ListPricingByTier(r.Context(), ref)
		if err != nil {
			if err == sql.ErrNoRows {
				w.WriteHeader(http.StatusNotFound)
				return
			}
			fmt.Fprintf(os.Stderr, "service ListPricingByTier(ctx, %s) error: %+v", ref, err)
			return
		}
		w.WriteHeader(http.StatusOK) // 200 OK
		json.NewEncoder(w).Encode(pmap)
	}
}
