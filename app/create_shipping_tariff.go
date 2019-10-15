package app

import (
	"encoding/json"
	"net/http"

	service "bitbucket.org/andyfusniakteam/ecom-api-go/service/firebase"
	log "github.com/sirupsen/logrus"
)

type createShippingTariffRequestBody struct {
	CountryCode  string `json:"country_code"`
	ShippingCode string `json:"shipping_code"`
	Name         string `json:"name"`
	Price        int    `json:"price"`
	TaxCode      string `json:"tax_code"`
}

func validateCreateShippingTariffRequest(request *createShippingTariffRequestBody) (bool, string) {
	if request.CountryCode == "" {
		return false, "attribute country_code must be set"
	}

	if request.ShippingCode == "" {
		return false, "attribute shipping_code must be set"
	}

	if request.Name == "" {
		return false, "attribute name must be set"
	}

	if request.Price < 0 {
		return false, "attribute price must be a positive integer"
	}

	if request.TaxCode == "" {
		return false, "attribute tax_code must be set"
	}
	return true, ""
}

// CreateShippingTariffHandler creates a shipping tariff
func (a *App) CreateShippingTariffHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		contextLogger := log.WithContext(ctx)
		contextLogger.Info("app: CreateShippingTariffHandler called")

		// parse the request body
		if r.Body == nil {
			// 400 Bad Request
			clientError(w, http.StatusBadRequest, ErrCodeBadRequest, "missing request body")
			return
		}
		request := createShippingTariffRequestBody{}
		dec := json.NewDecoder(r.Body)
		dec.DisallowUnknownFields()
		err := dec.Decode(&request)
		if err != nil {
			clientError(w, http.StatusBadRequest, ErrCodeBadRequest, err.Error())
			return
		}
		defer r.Body.Close()

		valid, message := validateCreateShippingTariffRequest(&request)
		if !valid {
			clientError(w, http.StatusBadRequest, ErrCodeBadRequest, message)
			return
		}

		// attempt to create the shipping tariff
		tariff, err := a.Service.CreateShippingTariff(ctx, request.CountryCode, request.ShippingCode, request.Name, request.Price, request.TaxCode)
		if err == service.ErrShippingTariffCodeExists {
			clientError(w, http.StatusConflict, ErrCodeShippingTariffCodeExists, "shipping tariff code already exists")
			return
		}
		if err != nil {
			contextLogger.Errorf("app: a.Service.CreateShippingTariff(ctx, countryCode=%q, shippingcode=%q, name=%q, price=%d, taxCode=%q) failed: %+v", request.CountryCode, request.ShippingCode, request.Name, request.Price, request.TaxCode, err)
			w.WriteHeader(http.StatusInternalServerError) // 500
			return
		}

		w.WriteHeader(http.StatusOK) // 200 OK
		json.NewEncoder(w).Encode(&tariff)
	}
}
