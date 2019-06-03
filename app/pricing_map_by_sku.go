package app

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/go-chi/chi"
)

// PricingMapBySKUHandler creates a handler function that returns a
// map of tier refs to PricingEntry.
func (a *App) PricingMapBySKUHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		sku := chi.URLParam(r, "sku")
		pmap, err := a.Service.PricingMapBySKU(r.Context(), sku)
		if err != nil {
			if err == sql.ErrNoRows {
				w.WriteHeader(http.StatusNotFound)
				return
			}
			fmt.Fprintf(os.Stderr, "service PricingMapBySKU(ctx, %s) error: %+v", sku, err)
			w.WriteHeader(http.StatusInternalServerError) // 500 Internal Server Error
			return
		}
		w.WriteHeader(http.StatusOK) // 200 OK
		json.NewEncoder(w).Encode(pmap)
	}
}
