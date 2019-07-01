package app

import (
	"encoding/json"
	"net/http"

	log "github.com/sirupsen/logrus"
)

// GetCatalogAssocsHandler creates a handler to return the entire catalog
func (app *App) GetCatalogAssocsHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		contextLogger := log.WithContext(ctx)
		contextLogger.Info("App: GetCatalogAssocsHandler started")

		cpo, err := app.Service.GetCategoryAssocs(ctx)
		if err != nil {
			contextLogger.Errorf("service GetCategoryAssocs(ctx) error: %v", err)
			w.WriteHeader(http.StatusInternalServerError) // 500
			json.NewEncoder(w).Encode(struct {
				Status  int    `json:"status"`
				Code    string `json:"code"`
				Message string `json:"message"`
			}{
				http.StatusInternalServerError,
				ErrCodeInternalServerError,
				err.Error(),
			})
			return
		}
		w.WriteHeader(http.StatusOK) // 200 OK
		json.NewEncoder(w).Encode(cpo)
	}
}
