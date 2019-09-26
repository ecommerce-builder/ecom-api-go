package app

import (
	"encoding/json"
	"net/http"

	service "bitbucket.org/andyfusniakteam/ecom-api-go/service/firebase"
	"github.com/go-chi/chi"
	log "github.com/sirupsen/logrus"
)

type updateShippingTariffRequest struct {
	ID           string `json:"id"`
	CountryCode  string `json:"country_code"`
	ShippingCode string `json:"shipping_code"`
	Name         string `json:"name"`
	Price        int    `json:"price"`
	TaxCode      string `json:"tax_code"`
}

// UpdateShippingTariffHandler creates a handler function that updates
// a shipping tariff with the given id.
func (a *App) UpdateShippingTariffHandler() http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		contextLogger := log.WithContext(ctx)
		contextLogger.Info("app: UpdateShippingTariffHandler started")

		request := createShippingTariffRequestBody{}
		dec := json.NewDecoder(r.Body)
		dec.DisallowUnknownFields()
		if err := dec.Decode(&request); err != nil {
			clientError(w, http.StatusBadRequest, ErrCodeBadRequest, err.Error())
			return
		}

		ok, message := validateCreateShippingTariffRequest(&request)
		if !ok {
			clientError(w, http.StatusBadRequest, ErrCodeBadRequest, message)
			return
		}

		shippingTariffID := chi.URLParam(r, "id")
		shippingTariff, err := a.Service.UpdateShippingTariff(ctx, shippingTariffID, request.CountryCode, request.ShippingCode, request.Name, request.Price, request.TaxCode)
		if err != nil {
			if err == service.ErrShippingTariffNotFound {
				// 404 Not Found
				clientError(w, http.StatusNotFound, ErrCodeShippingTariffNotFound, "shipping tariff code not found")
				return
			} else if err == service.ErrShippingTariffCodeExists {
				// 409 Conflict
				clientError(w, http.StatusConflict, ErrCodeShippingTariffCodeExists, "shipping tariff code already exists")
				return
			}

			contextLogger.Errorf("app: a.Service.UpdateShippingTariff(ctx, shippingTariffID=%q, countryCode=%q, shippingCode=%q, name=%q, price=%d, taxCode=%q) failed: %+v", shippingTariffID, request.CountryCode, request.ShippingCode, request.Name, request.Price, request.TaxCode, err)
			w.WriteHeader(http.StatusInternalServerError) // 500 Internal Server Error
			return
		}
		w.WriteHeader(http.StatusOK) // 200 OK
		json.NewEncoder(w).Encode(&shippingTariff)
	}
}
