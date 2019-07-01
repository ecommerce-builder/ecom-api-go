package app

import (
	"net/http"

	log "github.com/sirupsen/logrus"
)

// PurgeCatalogAssocsHandler deletes all catalog product associations.
func (a *App) PurgeCatalogAssocsHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		contextLogger := log.WithContext(ctx)
		contextLogger.Info("App: PurgeCatalogAssocsHandler started")

		_, err := a.Service.DeleteCategoryAssocs(ctx)
		if err != nil {
			contextLogger.Errorf("service DeleteCategoryAssocs(ctx) error: %+v", err)
			w.WriteHeader(http.StatusInternalServerError) // 500 Internal Server Error
			return
		}
		w.Header().Del("Content-Type")
		w.WriteHeader(http.StatusNoContent) // 204 No Content
	}
}
