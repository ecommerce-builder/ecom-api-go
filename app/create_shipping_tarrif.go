package app

import (
	"encoding/json"
	"net/http"

	service "bitbucket.org/andyfusniakteam/ecom-api-go/service/firebase"
	log "github.com/sirupsen/logrus"
)

type createShippingTarrifRequestBody struct {
	CountryCode  string `json:"country_code"`
	ShippingCode string `json:"shipping_code"`
	Name         string `json:"name"`
	Price        int    `json:"price"`
	TaxCode      string `json:"tax_code"`
}

func validateCreateShippingTarrifRequest(request *createShippingTarrifRequestBody) (bool, string) {
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

// CreateShippingTarrifHandler creates a shipping tarrif
func (a *App) CreateShippingTarrifHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		contextLogger := log.WithContext(ctx)
		contextLogger.Info("app: CreateShippingTarrifHandler called")

		// parse the request body
		if r.Body == nil {
			w.WriteHeader(http.StatusBadRequest) // 400 Bad Request
			clientError(w, http.StatusBadRequest, ErrCodeBadRequest, "missing request body")
			return
		}
		request := createShippingTarrifRequestBody{}
		err := json.NewDecoder(r.Body).Decode(&request)
		if err != nil {
			clientError(w, http.StatusBadRequest, ErrCodeBadRequest, err.Error())
			return
		}
		defer r.Body.Close()

		valid, message := validateCreateShippingTarrifRequest(&request)
		if !valid {
			clientError(w, http.StatusBadRequest, ErrCodeBadRequest, message)
			return
		}

		// attempt to create the shipping tarrif
		tarrif, err := a.Service.CreateShippingTarrif(ctx, request.CountryCode, request.ShippingCode, request.Name, request.Price, request.TaxCode)
		if err != nil {
			if err == service.ErrShippingTarrifCodeExists {
				clientError(w, http.StatusConflict, ErrCodeShippingTarrifCodeExists, "shipping tarrif code already exists")
				return
			}
			contextLogger.Errorf("app: a.Service.CreateShippingTarrif(ctx, countryCode=%q, shippingcode=%q, name=%q, price=%d, taxCode=%q) failed: %+v", request.CountryCode, request.ShippingCode, request.Name, request.Price, request.TaxCode, err)
			w.WriteHeader(http.StatusInternalServerError) // 500
			return
		}

		w.WriteHeader(http.StatusOK) // 200 OK
		json.NewEncoder(w).Encode(&tarrif)
	}
}
