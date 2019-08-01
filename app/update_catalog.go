package app

import (
	"encoding/json"
	"fmt"
	"net/http"

	"bitbucket.org/andyfusniakteam/ecom-api-go/service/firebase"
	log "github.com/sirupsen/logrus"
)

// UpdateCatalogHandler creates an HTTP handler that updates the catalog.
func (a *App) UpdateCatalogHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		contextLogger := log.WithContext(ctx)
		contextLogger.Info("App: UpdateCatalogHandler started")

		cats := firebase.Category{}
		if err := json.NewDecoder(r.Body).Decode(&cats); err != nil {
			w.WriteHeader(http.StatusBadRequest) // 400 Bad Request
			json.NewEncoder(w).Encode(struct {
				Status  int    `json:"status"`
				Code    string `json:"code"`
				Message string `json:"message"`
			}{
				http.StatusBadRequest,
				ErrCodeBadRequest,
				err.Error(),
			})
			return
		}
		defer r.Body.Close()
		if err := a.Service.UpdateCatalog(ctx, &cats); err != nil {
			if err == firebase.ErrAssocsAlreadyExist {
				w.WriteHeader(http.StatusConflict) // 409 Conflict
				json.NewEncoder(w).Encode(struct {
					Status  int    `json:"status"`
					Code    string `json:"code"`
					Message string `json:"message"`
				}{
					http.StatusConflict,
					ErrCodeAssocsAlreadyExist,
					fmt.Sprintf("catalog associations already exist"),
				})
				return
			}
			contextLogger.Errorf("UpdateCatalog(ctx, cats) failed: %+v", err)
			w.WriteHeader(http.StatusInternalServerError) // 500 Internal Server Error
			return
		}
		w.Header().Del("Content-Type")
		w.WriteHeader(http.StatusNoContent) // 204 No Content
	}
}
