package app

import (
	"encoding/json"
	"net/http"

	service "bitbucket.org/andyfusniakteam/ecom-api-go/service/firebase"
	log "github.com/sirupsen/logrus"
)

// GetProductPrices creates a handler function that returns a
// list of prices.
func (a *App) GetProductPrices() http.HandlerFunc {
	type response struct {
		Object string           `json:"object"`
		Data   []*service.Price `json:"data"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		contextLogger := log.WithContext(ctx)
		contextLogger.Info("app: GetProductPrices started")

		productID := r.URL.Query().Get("product_id")
		priceListID := r.URL.Query().Get("price_list_id")

		// TODO: validate query params

		prices, err := a.Service.GetPrices(ctx, productID, priceListID)
		if err == service.ErrProductNotFound {
			clientError(w, http.StatusNotFound, ErrCodeProductNotFound,
				"product not found") // 404
			return
		}
		if err == service.ErrPriceListNotFound {
			clientError(w, http.StatusNotFound, ErrCodePriceListNotFound,
				"price list not found") // 404
			return
		}
		if err != nil {
			contextLogger.Errorf("app: GetPrices(ctx, productID=%q, priceListID=%q) failed: %+v", productID, priceListID, err)
			w.WriteHeader(http.StatusInternalServerError) // 500
			return
		}

		list := response{
			Object: "list",
			Data:   prices,
		}
		w.WriteHeader(http.StatusOK) // 200
		json.NewEncoder(w).Encode(&list)
	}
}
