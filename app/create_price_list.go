package app

import (
	"encoding/json"
	"net/http"

	service "bitbucket.org/andyfusniakteam/ecom-api-go/service/firebase"
	log "github.com/sirupsen/logrus"
)

// CreatePriceListHandler creates a new product
func (a *App) CreatePriceListHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		contextLogger := log.WithContext(ctx)
		contextLogger.Info("app: CreatePriceListHandler called")

		requestBody := service.PriceListCreate{}
		dec := json.NewDecoder(r.Body)
		dec.DisallowUnknownFields()
		if err := dec.Decode(&requestBody); err != nil {
			clientError(w, http.StatusBadRequest,
				ErrCodeBadRequest, err.Error()) // 400
			return
		}

		valid, message := validateCreatePriceListRequest(&requestBody)
		if !valid {
			clientError(w, http.StatusBadRequest,
				ErrCodeBadRequest, message) // 400
			return
		}

		priceList, err := a.Service.CreatePriceList(ctx, &requestBody)
		if err == service.ErrPriceListCodeExists {
			clientError(w, http.StatusConflict,
				ErrCodePriceListCodeExists,
				"price list is already in use") // 409
			return
		}
		if err != nil {
			contextLogger.Errorf("app: a.Service.CreatePriceList(ctx, &requestBody) failed: %+v", err)
			w.WriteHeader(http.StatusInternalServerError) // 500
			return
		}
		w.WriteHeader(http.StatusCreated) // 201
		json.NewEncoder(w).Encode(priceList)
	}
}

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
