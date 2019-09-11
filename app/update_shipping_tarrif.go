package app

import (
	"encoding/json"
	"net/http"

	service "bitbucket.org/andyfusniakteam/ecom-api-go/service/firebase"
	"github.com/go-chi/chi"
	log "github.com/sirupsen/logrus"
)

type updateShippingTarrifRequest struct {
	ID           string `json:"id"`
	CountryCode  string `json:"country_code"`
	ShippingCode string `json:"shipping_code"`
	Name         string `json:"name"`
	Price        int    `json:"price"`
	TaxCode      string `json:"tax_code"`
}

// UpdateShippingTarrifHandler creates a handler function that updates
// a shipping tarrif with the given id.
func (a *App) UpdateShippingTarrifHandler() http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		contextLogger := log.WithContext(ctx)
		contextLogger.Info("app: UpdateShippingTarrifHandler started")

		request := createShippingTarrifRequestBody{}
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			clientError(w, http.StatusBadRequest, ErrCodeBadRequest, err.Error())
			return
		}

		ok, message := validateCreateShippingTarrifRequest(&request)
		if !ok {
			clientError(w, http.StatusBadRequest, ErrCodeBadRequest, message)
			return
		}

		shippingTarrifID := chi.URLParam(r, "id")
		shippingTarrif, err := a.Service.UpdateShippingTarrif(ctx, shippingTarrifID, request.CountryCode, request.ShippingCode, request.Name, request.Price, request.TaxCode)
		if err != nil {
			if err == service.ErrShippingTarrifNotFound {
				// 404 Not Found
				clientError(w, http.StatusNotFound, ErrCodeShippingTarrifNotFound, "shipping tarrif code not found")
				return
			} else if err == service.ErrShippingTarrifCodeExists {
				// 409 Conflict
				clientError(w, http.StatusConflict, ErrCodeShippingTarrifCodeExists, "shipping tarrif code already exists")
				return
			}

			contextLogger.Errorf("app: a.Service.UpdateShippingTarrif(ctx, shippingTarrifID=%q, countryCode=%q, shippingCode=%q, name=%q, price=%d, taxCode=%q) failed: %+v", shippingTarrifID, request.CountryCode, request.ShippingCode, request.Name, request.Price, request.TaxCode, err)
			w.WriteHeader(http.StatusInternalServerError) // 500 Internal Server Error
			return
		}
		w.WriteHeader(http.StatusOK) // 200 OK
		json.NewEncoder(w).Encode(&shippingTarrif)
	}
}
