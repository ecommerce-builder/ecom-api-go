package app

import (
	"encoding/json"
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
		contextLogger.Info("App: DeletePriceListHandler started")

		priceListID := chi.URLParam(r, "id")
		if err := a.Service.DeletePriceList(ctx, priceListID); err != nil {
			if err == service.ErrPriceListNotFound {
				w.WriteHeader(http.StatusNotFound) // 404 Not Found
				return
			} else if err == service.ErrPriceListInUse {
				w.WriteHeader(http.StatusConflict) // 409 Conflict
				json.NewEncoder(w).Encode(struct {
					Status  int    `json:"status"`
					Code    string `json:"code"`
					Message string `json:"message"`
				}{
					http.StatusConflict,
					ErrCodePriceListInUse,
					"price list is already in use",
				})
				return
			}
			contextLogger.Errorf("app: a.Service.DeletePriceList(ctx, priceListID=%q) error: %+v", priceListID, err)
			w.WriteHeader(http.StatusInternalServerError) // 500 Internal Server Error
			return
		}
		w.Header().Del("Content-Type")
		w.WriteHeader(http.StatusNoContent) // 204 No Content
	}
}
