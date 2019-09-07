package app

import (
	"encoding/json"
	"net/http"

	service "bitbucket.org/andyfusniakteam/ecom-api-go/service/firebase"
	"github.com/go-chi/chi"
	log "github.com/sirupsen/logrus"
)

// UpdatePriceListHandler creates a handler function that updates
// as price list with the given price list id.
func (a *App) UpdatePriceListHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		contextLogger := log.WithContext(ctx)
		contextLogger.Info("app: UpdatePriceListHandler started")

		requestBody := service.PriceListCreate{}
		if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
			clientError(w, http.StatusBadRequest, ErrCodeBadRequest, err.Error())
			return
		}

		ok, message := validateCreatePriceListRequest(&requestBody)
		if !ok {
			clientError(w, http.StatusBadRequest, ErrCodeBadRequest, message)
			return
		}

		priceListID := chi.URLParam(r, "id")
		priceList, err := a.Service.UpdatePriceList(ctx, priceListID, &requestBody)
		if err != nil {
			if err == service.ErrPriceListNotFound {
				// 404 Not Found
				clientError(w, http.StatusNotFound, ErrCodePriceListNotFound, "price list not found")
				return
			} else if err == service.ErrPriceListCodeExists {
				// 409 Conflict
				clientError(w, http.StatusConflict, ErrCodePriceListCodeExists, "price list is already in use")
				return
			}
			contextLogger.Errorf("app: a.Service.UpdatePriceList(ctx, priceListID=%q, p=%v) %+v", priceListID, requestBody, err)
			w.WriteHeader(http.StatusInternalServerError) // 500 Internal Server Error
			return
		}
		w.WriteHeader(http.StatusOK) // 200 OK
		json.NewEncoder(w).Encode(priceList)
	}
}
