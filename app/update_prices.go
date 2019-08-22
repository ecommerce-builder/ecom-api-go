package app

import (
	"encoding/json"
	"net/http"

	service "bitbucket.org/andyfusniakteam/ecom-api-go/service/firebase"
	"github.com/go-chi/chi"
	log "github.com/sirupsen/logrus"
)

type pricingResponseBody struct {
	Object string `json:"object"`
	*service.Price
}

// UpdateTierPricingHandler creates a handler function that updates
// as tier price for a product of SKU with tier ref.
func (a *App) UpdateTierPricingHandler() http.HandlerFunc {
	type updateTierPricingRequest struct {
		UnitPrice float64 `json:"unit_price"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		contextLogger := log.WithContext(ctx)
		contextLogger.Info("App: UpdateTierPricingHandler started")

		sku := chi.URLParam(r, "sku")
		ref := chi.URLParam(r, "ref")
		var req updateTierPricingRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			w.WriteHeader(http.StatusBadRequest) // 400 Bad Request
			json.NewEncoder(w).Encode(struct {
				Status  int    `json:"status"`
				Code    string `json:"code"`
				Message string `json:"message"`
			}{
				http.StatusBadRequest,
				ErrCodeBadRequest,
				err.Error(),
			})
			return
		}
		if ref != "default" {
			w.WriteHeader(http.StatusConflict)
			return
		}
		price, err := a.Service.UpdateTierPricing(ctx, sku, ref, req.UnitPrice)
		if err != nil {
			if err == service.ErrPriceListNotFound {
				w.WriteHeader(http.StatusNotFound) // 404 Not Found
				json.NewEncoder(w).Encode(struct {
					Status  int    `json:"status"`
					Code    string `json:"code"`
					Message string `json:"message"`
				}{
					http.StatusNotFound,
					ErrCodePriceListNotFound,
					"price list not found",
				})
				return
			} else if err == service.ErrProductSKUExists {
				w.WriteHeader(http.StatusConflict) // 409 Conflict
				json.NewEncoder(w).Encode(struct {
					Status  int    `json:"status"`
					Code    string `json:"code"`
					Message string `json:"message"`
				}{
					http.StatusConflict,
					ErrCodePriceListCodeExists,
					"price sku already exists",
				})
				return
			}
			contextLogger.Errorf("service UpdateTierPricing(ctx, %s, %s) error: %+v", sku, ref, err)
			w.WriteHeader(http.StatusInternalServerError) // 500 Internal Server Error
			return
		}

		res := pricingResponseBody{
			Object: "pricing",
			Price:  price,
		}
		w.WriteHeader(http.StatusOK) // 200 OK
		json.NewEncoder(w).Encode(res)
	}
}
