package app

import (
	"encoding/json"
	"net/http"

	service "bitbucket.org/andyfusniakteam/ecom-api-go/service/firebase"
	log "github.com/sirupsen/logrus"
)

func validateCreatePriceListRequest(requestBody *service.PriceListCreate) (bool, string) {
	if requestBody.PriceListCode == "" {
		return false, "price_list_code attribute is required"
	}

	if len(requestBody.PriceListCode) < 3 || len(requestBody.PriceListCode) > 16 {
		return false, "price_list_code attribute must be between 3 and 16 characters in length"
	}

	if requestBody.CurrencyCode != "GBP" &&
		requestBody.CurrencyCode != "EUR" &&
		requestBody.CurrencyCode != "USD" {
		return false, "currency_code attribute must be a value of GBP, EUR or USD"
	}

	if requestBody.Strategy != "simple" &&
		requestBody.Strategy != "volume" &&
		requestBody.Strategy != "tiered" {
		return false, "strategy attribute must be a value of simple, volume or tiered"
	}

	return true, ""
}

// CreatePriceListHandler creates a new product
func (a *App) CreatePriceListHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		contextLogger := log.WithContext(ctx)
		contextLogger.Info("App: CreatePriceListHandler called")

		requestBody := service.PriceListCreate{}
		if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
			http.Error(w, err.Error(), 400)
			return
		}

		valid, message := validateCreatePriceListRequest(&requestBody)
		if !valid {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(struct {
				Status  int    `json:"status"`
				Code    string `json:"code"`
				Message string `json:"message"`
			}{
				http.StatusBadRequest,
				ErrCodeBadRequest,
				message,
			})
			return
		}

		priceList, err := a.Service.CreatePriceList(ctx, &requestBody)
		if err != nil {
			if err == service.ErrPriceListCodeTaken {
				w.WriteHeader(http.StatusConflict) // 409 Conflict
				json.NewEncoder(w).Encode(struct {
					Status  int    `json:"status"`
					Code    string `json:"code"`
					Message string `json:"message"`
				}{
					http.StatusConflict,
					ErrCodePriceListCodeTaken,
					"price list is already in use",
				})
				return
			}
		}
		w.WriteHeader(http.StatusCreated) // 201 Created
		json.NewEncoder(w).Encode(priceList)
	}
}
