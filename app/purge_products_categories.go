package app

import (
	"net/http"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

// DeleteProductsCategoriesHandler returns a handler to purge all product to categories.
func (app *App) DeleteProductsCategoriesHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		contextLogger := log.WithContext(ctx)
		contextLogger.Info("app: DeleteProductsCategoriesHandler called")

		err := app.Service.DeleteAllProductCategoryRelations(ctx)
		if err != nil {
			contextLogger.Errorf("app: DeleteProductsCategories(ctx) error: %+v", errors.WithStack(err))
			w.WriteHeader(http.StatusInternalServerError) // 500
			return
		}

		w.Header().Del("Content-Type")
		w.Header().Set("Content-Length", "0")
		w.WriteHeader(http.StatusNoContent) // 204 No Content
	}
}
