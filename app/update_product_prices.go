package app

import (
	"encoding/json"
	"net/http"

	service "bitbucket.org/andyfusniakteam/ecom-api-go/service/firebase"
	log "github.com/sirupsen/logrus"
)

// UpdateProductPricesHandler creates a handler function that updates
// as tier price for a product of SKU with tier ref.
func (a *App) UpdateProductPricesHandler() http.HandlerFunc {
	type listPricesResponse struct {
		Object string           `json:"object"`
		Data   []*service.Price `json:"data"`
	}

	type updatePriceRequest struct {
		Object string                  `json:"object"`
		Data   []*service.PriceRequest `json:"data"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		contextLogger := log.WithContext(ctx)
		contextLogger.Info("app: UpdateProductPricesHandler started")

		productID := r.URL.Query().Get("product_id")
		priceListID := r.URL.Query().Get("price_list_id")
		contextLogger.Debugf("app: query params product_id=%q, price_list_id=%q", productID, priceListID)

		var request updatePriceRequest
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			clientError(w, http.StatusBadRequest, ErrCodeBadRequest, err.Error())
			return
		}

		prices, err := a.Service.UpdateProductPrices(ctx, productID, priceListID, request.Data)
		if err == service.ErrProductNotFound {
			clientError(w, http.StatusNotFound, ErrCodeProductNotFound, "product not found")
			return
		}
		if err == service.ErrPriceListNotFound {
			clientError(w, http.StatusNotFound, ErrCodePriceListNotFound, "price list not found")
			return
		}
		if err != nil {
			contextLogger.Errorf("app: UpdateProductPrices(ctx, productID=%q, priceListID=%q, request=%v) failed: %+v", productID, priceListID, request, err)
			w.WriteHeader(http.StatusInternalServerError) // 500 Internal Server Error
			return
		}

		list := listPricesResponse{
			Object: "list",
			Data:   prices,
		}
		w.WriteHeader(http.StatusCreated) // 200 Created
		json.NewEncoder(w).Encode(&list)
	}
}
