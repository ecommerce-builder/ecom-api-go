package app

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"bitbucket.org/andyfusniakteam/ecom-api-go/service/firebase"
	"github.com/go-chi/chi"
)

// UpdateTierPricingHandler creates a handler function that updates
// as tier price for a product of SKU with tier ref.
func (a *App) UpdateTierPricingHandler() http.HandlerFunc {
	type updateTierPricingRequest struct {
		UnitPrice float64 `json:"unit_price"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		sku := chi.URLParam(r, "sku")
		ref := chi.URLParam(r, "ref")
		var req updateTierPricingRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			w.WriteHeader(http.StatusBadRequest) // 400 Bad Request
			json.NewEncoder(w).Encode(struct {
				Code    int    `json:"code"`
				Message string `json:"message"`
			}{
				400,
				err.Error(),
			})
			return
		}
		if ref != "default" {
			w.WriteHeader(http.StatusConflict)
			return
		}
		pricing, err := a.Service.UpdateTierPricing(r.Context(), sku, ref, req.UnitPrice)
		if err != nil {
			if err == firebase.ErrTierPricingNotFound {
				w.WriteHeader(http.StatusNotFound) // 404 Not Found
				return
			}
			w.WriteHeader(http.StatusInternalServerError) // 500 Internal Server Error
			fmt.Fprintf(os.Stderr, "service UpdateTierPricing(ctx, %s, %s) error: %+v", sku, ref, err)
			return
		}
		fmt.Println(pricing)
		w.WriteHeader(http.StatusOK) // 200 OK
		json.NewEncoder(w).Encode(*pricing)
	}
}
