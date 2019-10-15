package app

import (
	"net/http"

	service "bitbucket.org/andyfusniakteam/ecom-api-go/service/firebase"
	"github.com/go-chi/chi"
	log "github.com/sirupsen/logrus"
)

// DeletePriceListHandler creates a handler function that deletes
// a price list by id.
func (a *App) DeletePriceListHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		contextLogger := log.WithContext(ctx)
		contextLogger.Info("app: DeletePriceListHandler started")

		priceListID := chi.URLParam(r, "id")
		err := a.Service.DeletePriceList(ctx, priceListID)
		if err == service.ErrPriceListNotFound {
			clientError(w, http.StatusNotFound, ErrCodePriceListNotFound, "price list not found")
			return
		}
		if err == service.ErrPriceListInUse {
			clientError(w, http.StatusConflict, ErrCodePriceListInUse, "price list is already in use")
			return
		}
		if err != nil {
			contextLogger.Errorf("app: a.Service.DeletePriceList(ctx, priceListID=%q) error: %+v", priceListID, err)
			w.WriteHeader(http.StatusInternalServerError) // 500 Internal Server Error
			return
		}
		w.Header().Del("Content-Type")
		w.WriteHeader(http.StatusNoContent) // 204 No Content
	}
}
