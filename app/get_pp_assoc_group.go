package app

import (
	"encoding/json"
	"net/http"

	service "bitbucket.org/andyfusniakteam/ecom-api-go/service/firebase"
	"github.com/go-chi/chi"
	log "github.com/sirupsen/logrus"
)

// GetPPAssocGroupHandler creates a handler function that returns a
// single product to product association group.
func (a *App) GetPPAssocGroupHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		contextLogger := log.WithContext(ctx)
		contextLogger.Info("app: GetPPAssocGroupHandler called")

		ppAssocGroupID := chi.URLParam(r, "id")
		if !IsValidUUID(ppAssocGroupID) {
			clientError(w, http.StatusBadRequest, ErrCodeBadRequest,
				"URL parameter id must be a valid v4 UUID") // 400
			return
		}
		ppAssocGroup, err := a.Service.GetPPAssocGroup(ctx, ppAssocGroupID)
		if err == service.ErrPPAssocGroupNotFound {
			clientError(w, http.StatusNotFound, ErrCodePPAssocGroupNotFound,
				"product to product association group not found") // 404
			return
		}
		if err != nil {
			contextLogger.Errorf("app: a.Service.GetPPAssocGroup(ctx, ppAssocGroupID=%q)", ppAssocGroupID)
			w.WriteHeader(http.StatusInternalServerError) // 500
			return
		}
		w.WriteHeader(http.StatusOK) // 200 OK
		json.NewEncoder(w).Encode(&ppAssocGroup)
	}
}
