package app

import (
	"net/http"

	service "bitbucket.org/andyfusniakteam/ecom-api-go/service/firebase"
	"github.com/go-chi/chi"
	log "github.com/sirupsen/logrus"
)

// DeleteShippingTarrifHandler create a handler to delete a shipping tarrif.
func (a *App) DeleteShippingTarrifHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		contextLogger := log.WithContext(ctx)
		contextLogger.Info("app: DeleteShippingTarrifHandler started")

		shippingTarrifID := chi.URLParam(r, "id")
		if err := a.Service.DeleteShippingTarrif(ctx, shippingTarrifID); err != nil {
			if err == service.ErrShippingTarrifNotFound {
				clientError(w, http.StatusNotFound, ErrCodeShippingTarrifNotFound, "shipping tarrif not found")
				return
			}
			contextLogger.Errorf("app DeleteShippingTarrif(ctx, shippingTarrifID=%q) failed: %+v", shippingTarrifID, err)
			w.WriteHeader(http.StatusInternalServerError) // 500 Internal Server Error
			return
		}
		w.Header().Del("Content-Type")
		w.Header().Set("Content-Length", "0")
		w.WriteHeader(http.StatusNoContent) // 204 No Content
	}
}
