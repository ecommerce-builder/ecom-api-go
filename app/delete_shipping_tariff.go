package app

import (
	"net/http"

	service "bitbucket.org/andyfusniakteam/ecom-api-go/service/firebase"
	"github.com/go-chi/chi"
	log "github.com/sirupsen/logrus"
)

// DeleteShippingTariffHandler create a handler to delete a shipping tariff.
func (a *App) DeleteShippingTariffHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		contextLogger := log.WithContext(ctx)
		contextLogger.Info("app: DeleteShippingTariffHandler started")

		shippingTariffID := chi.URLParam(r, "id")
		err := a.Service.DeleteShippingTariff(ctx, shippingTariffID)
		if err == service.ErrShippingTariffNotFound {
			clientError(w, http.StatusNotFound, ErrCodeShippingTariffNotFound, "shipping tariff not found")
			return
		}
		if err != nil {
			contextLogger.Errorf("app: DeleteShippingTariff(ctx, shippingTariffID=%q) failed: %+v", shippingTariffID, err)
			w.WriteHeader(http.StatusInternalServerError) // 500 Internal Server Error
			return
		}
		w.Header().Del("Content-Type")
		w.Header().Set("Content-Length", "0")
		w.WriteHeader(http.StatusNoContent) // 204 No Content
	}
}
