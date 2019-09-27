package app

import (
	"encoding/json"
	"net/http"

	service "bitbucket.org/andyfusniakteam/ecom-api-go/service/firebase"
	"github.com/go-chi/chi"
	log "github.com/sirupsen/logrus"
)

// GetPPAssocHandler creates a handler function that returns a
// product to product association
func (a *App) GetPPAssocHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		contextLogger := log.WithContext(ctx)
		contextLogger.Info("app: GetPPAssocHandler called")

		ppAssocID := chi.URLParam(r, "id")
		if !IsValidUUID(ppAssocID) {
			clientError(w, http.StatusBadRequest, ErrCodeBadRequest, "URL parameter id must be a valid v4 UUID")
			return
		}

		ppAssoc, err := a.Service.GetPPAssoc(ctx, ppAssocID)
		if err != nil {
			if err == service.ErrPPAssocNotFound {
				clientError(w, http.StatusNotFound, ErrCodePPAssocNotFound, "product to product association not found")
				return
			}
			contextLogger.Errorf("app: a.Service.GetPPAssoc(ctx, ppAssocID=%q) failed: %+v", ppAssocID, err)
			w.WriteHeader(http.StatusInternalServerError) // 500 Internal Server Error
			return
		}
		w.WriteHeader(http.StatusOK) // 200 OK
		json.NewEncoder(w).Encode(&ppAssoc)
	}
}
