package app

import (
	"encoding/json"
	"net/http"

	service "bitbucket.org/andyfusniakteam/ecom-api-go/service/firebase"
	log "github.com/sirupsen/logrus"
)

type createShippingTariffRequestBody struct {
	CountryCode  *string `json:"country_code"`
	ShippingCode *string `json:"shipping_code"`
	Name         *string `json:"name"`
	Price        *int    `json:"price"`
	TaxCode      *string `json:"tax_code"`
}

func validateCreateShippingTariffRequest(request *createShippingTariffRequestBody) (bool, string) {
	// country_code attribute
	if request.CountryCode == nil {
		return false, "attribute country_code must be set"
	}

	// shipping_code attribute
	if request.ShippingCode == nil {
		return false, "attribute shipping_code must be set"
	}

	// name attribute
	if request.Name == nil {
		return false, "attribute name must be set"
	}

	// price attribute
	if request.Price == nil {
		return false, "attribute price must be set"
	}
	if *request.Price < 0 {
		return false, "attribute price must be a positive integer"
	}

	// tax_code attribute
	if request.TaxCode == nil {
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
			clientError(w, http.StatusBadRequest, ErrCodeBadRequest,
				"missing request body") // 400
			return
		}
		request := createShippingTariffRequestBody{}
		dec := json.NewDecoder(r.Body)
		dec.DisallowUnknownFields()
		err := dec.Decode(&request)
		if err != nil {
			clientError(w, http.StatusBadRequest, ErrCodeBadRequest,
				err.Error()) // 400
			return
		}
		defer r.Body.Close()

		valid, message := validateCreateShippingTariffRequest(&request)
		if !valid {
			clientError(w, http.StatusBadRequest, ErrCodeBadRequest,
				message) // 400
			return
		}

		// attempt to create the shipping tariff
		tariff, err := a.Service.CreateShippingTariff(ctx, *request.CountryCode,
			*request.ShippingCode, *request.Name, *request.Price, *request.TaxCode)
		if err == service.ErrShippingTariffCodeExists {
			clientError(w, http.StatusConflict, ErrCodeShippingTariffCodeExists,
				"shipping tariff code already exists") // 409
			return
		}
		if err != nil {
			contextLogger.Errorf("app: a.Service.CreateShippingTariff(ctx, countryCode=%q, shippingcode=%q, name=%q, price=%d, taxCode=%q) failed: %+v", request.CountryCode, request.ShippingCode, request.Name, request.Price, request.TaxCode, err)
			w.WriteHeader(http.StatusInternalServerError) // 500
			return
		}
		w.WriteHeader(http.StatusOK) // 200
		json.NewEncoder(w).Encode(&tariff)
	}
}
