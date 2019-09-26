package app

import (
	"net/http"

	service "bitbucket.org/andyfusniakteam/ecom-api-go/service/firebase"
	"github.com/go-chi/chi"
	log "github.com/sirupsen/logrus"
)

// DeletePPAssocHandler create a handler to delete a shipping tariff.
func (a *App) DeletePPAssocHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		contextLogger := log.WithContext(ctx)
		contextLogger.Info("app: DeletePPAssocHandler started")

		ppAssocID := chi.URLParam(r, "id")
		if !IsValidUUID(ppAssocID) {
			clientError(w, http.StatusBadRequest, ErrCodeBadRequest, "URL parameter id must be a valid v4 UUID")
			return
		}

		if err := a.Service.DeletePPAssoc(ctx, ppAssocID); err != nil {
			if err == service.ErrPPAssocNotFound {
				clientError(w, http.StatusNotFound, ErrCodeErrPPAssocNotFound, "product to product assoication not found")
				return
			}
			contextLogger.Errorf("app: a.Service.DeletePPAssoc(ctx, ppAssocID=%q) failed: %+v", ppAssocID, err)
			w.WriteHeader(http.StatusInternalServerError) // 500 Internal Server Error
			return
		}
		w.Header().Del("Content-Type")
		w.Header().Set("Content-Length", "0")
		w.WriteHeader(http.StatusNoContent) // 204 No Content
	}
}
