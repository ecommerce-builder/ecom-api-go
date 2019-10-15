package app

import (
	"encoding/json"
	"net/http"

	service "bitbucket.org/andyfusniakteam/ecom-api-go/service/firebase"
	"github.com/go-chi/chi"
	log "github.com/sirupsen/logrus"
)

// GetPriceListHandler creates a handler function that returns a
// price list.
func (a *App) GetPriceListHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		contextLogger := log.WithContext(ctx)
		contextLogger.Info("app: GetPriceListHandler called")

		priceListID := chi.URLParam(r, "id")
		priceList, err := a.Service.GetPriceList(ctx, priceListID)
		if err == service.ErrPriceListNotFound {
			clientError(w, http.StatusNotFound, ErrCodePriceListCodeExists,
				"price_list_id not found") // 404
			return
		}
		if err != nil {
			contextLogger.Errorf("app: a.Service.GetPriceList(ctx, priceListID=%q)", priceListID)
			w.WriteHeader(http.StatusInternalServerError) // 500
			return
		}
		w.WriteHeader(http.StatusOK) // 200
		json.NewEncoder(w).Encode(priceList)
	}
}
