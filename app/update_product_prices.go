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
		contextLogger.Info("App: UpdateProductPricesHandler started")

		productID := r.URL.Query().Get("product_id")
		priceListID := r.URL.Query().Get("price_list_id")

		var request updatePriceRequest
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
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

		prices, err := a.Service.UpdateProductPrices(ctx, productID, priceListID, request.Data)
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
			contextLogger.Errorf("app: UpdateProductPrices(ctx, productID, priceListID, request) failed: %+v", productID, priceListID, request, err)
			w.WriteHeader(http.StatusInternalServerError) // 500 Internal Server Error
			return
		}

		list := listPricesResponse{
			Object: "list",
			Data:   prices,
		}
		w.WriteHeader(http.StatusOK) // 200 OK
		json.NewEncoder(w).Encode(&list)
	}
}
