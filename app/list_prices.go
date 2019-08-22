package app

import (
	"encoding/json"
	"net/http"

	service "bitbucket.org/andyfusniakteam/ecom-api-go/service/firebase"
	log "github.com/sirupsen/logrus"
)

// ListPricesHandler creates a handler function that returns a
// list of prices.
func (a *App) ListPricesHandler() http.HandlerFunc {
	type listPriceListsResponse struct {
		Object string           `json:"object"`
		Data   []*service.Price `json:"data"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		contextLogger := log.WithContext(ctx)
		contextLogger.Info("App: ListPricesHandler started")

		productID := r.URL.Query().Get("product_id")
		priceListID := r.URL.Query().Get("price_list_id")
		priceLists, err := a.Service.GetPrices(ctx, productID, priceListID)
		if err != nil {
			contextLogger.Errorf("service GetPrices(ctx) error: %+v", err)
			w.WriteHeader(http.StatusInternalServerError) // 500 Internal Server Error
			return
		}

		list := listPriceListsResponse{
			Object: "list",
			Data:   priceLists,
		}
		w.WriteHeader(http.StatusOK) // 200 OK
		json.NewEncoder(w).Encode(&list)
	}
}
