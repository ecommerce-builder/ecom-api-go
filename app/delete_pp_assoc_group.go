package app

import (
	"net/http"

	service "bitbucket.org/andyfusniakteam/ecom-api-go/service/firebase"
	"github.com/go-chi/chi"
	log "github.com/sirupsen/logrus"
)

// DeletePPAssocGroupHandler create a handler to delete a shipping tariff.
func (a *App) DeletePPAssocGroupHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		contextLogger := log.WithContext(ctx)
		contextLogger.Info("app: DeletePPAssocGroupHandler started")

		ppAssocGroupID := chi.URLParam(r, "id")
		if err := a.Service.DeletePPAssocGroup(ctx, ppAssocGroupID); err != nil {
			if err == service.ErrPPAssocGroupNotFound {
				clientError(w, http.StatusNotFound, ErrCodePPAssocGroupNotFound, "product to product associations group not found")
				return
			} else if err == service.ErrPPAssocGroupContainsAssocs {
				clientError(w, http.StatusNotFound, ErrPPAssocGroupContainsAssocs, "product to product assocations group contains associations - delete them first")
				return
			}
			contextLogger.Errorf("app: a.Service.DeletePPAssocGroup(ctx, ppAssocGroupID=%q) failed: %+v", ppAssocGroupID, err)
			w.WriteHeader(http.StatusInternalServerError) // 500 Internal Server Error
			return
		}
		w.Header().Del("Content-Type")
		w.Header().Set("Content-Length", "0")
		w.WriteHeader(http.StatusNoContent) // 204 No Content
	}
}
