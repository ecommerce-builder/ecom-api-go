package app

import (
	"net/http"

	service "bitbucket.org/andyfusniakteam/ecom-api-go/service/firebase"
	"github.com/go-chi/chi"
	log "github.com/sirupsen/logrus"
)

// DeletePriceHandler creates a handler function that deletes
// a price list by id.
func (a *App) DeletePriceHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		contextLogger := log.WithContext(ctx)
		contextLogger.Info("app: DeletePriceListHandler started")

		priceListID := chi.URLParam(r, "id")
		if err := a.Service.DeletePriceList(ctx, priceListID); err != nil {
			if err == service.ErrPriceListNotFound {
				clientError(w, http.StatusNotFound, ErrCodePriceListNotFound, "price list not found")
				return
			}
			contextLogger.Errorf("app: DeletePriceList(ctx, productLidID=%q) error: %+v", priceListID, err)
			w.WriteHeader(http.StatusInternalServerError) // 500 Internal Server Error
			return
		}
		w.Header().Del("Content-Type")
		w.WriteHeader(http.StatusNoContent) // 204 No Content
	}
}
