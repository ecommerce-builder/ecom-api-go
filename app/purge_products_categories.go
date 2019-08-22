package app

import (
	"net/http"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

// PurgeProductsCategoriesHandler returns a handler to purge all product to categories.
func (app *App) PurgeProductsCategoriesHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		contextLogger := log.WithContext(ctx)
		contextLogger.Info("App: PurgeProductsCategoriesHandler called")

		err := app.Service.PurgeProductsCategories(ctx)
		if err != nil {
			contextLogger.Errorf("service PurgeProductsCategories(ctx) error: %+v", errors.WithStack(err))
			w.WriteHeader(http.StatusInternalServerError) // 500
			return
		}

		w.Header().Del("Content-Type")
		w.Header().Set("Content-Length", "0")
		w.WriteHeader(http.StatusNoContent) // 204 No Content
	}
}
