package app

import (
	"encoding/json"
	"net/http"

	service "bitbucket.org/andyfusniakteam/ecom-api-go/service/firebase"
	"github.com/go-chi/chi"
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
		contextLogger.Info("App: GetProductPrices started")

		productID := chi.URLParam(r, "id")
		priceListID := r.URL.Query().Get("price_list_id")

		prices, err := a.Service.GetPrices(ctx, productID, priceListID)
		if err != nil {
			if err == service.ErrProductNotFound {
				w.WriteHeader(http.StatusNotFound)
				json.NewEncoder(w).Encode(struct {
					Status  int    `json:"status"`
					Code    string `json:"code"`
					Message string `json:"message"`
				}{
					http.StatusNotFound,
					ErrCodeProductNotFound,
					"product not found",
				})
				return
			} else if err == service.ErrPriceListNotFound {
				w.WriteHeader(http.StatusNotFound)
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
			}
			contextLogger.Errorf("app: GetPrices(ctx, productID=%q, priceListID=%q) failed: %+v", productID, priceListID, err)
			w.WriteHeader(http.StatusInternalServerError) // 500 Internal Server Error
			return
		}

		list := response{
			Object: "list",
			Data:   prices,
		}
		w.WriteHeader(http.StatusOK) // 200 OK
		json.NewEncoder(w).Encode(&list)
	}
}
