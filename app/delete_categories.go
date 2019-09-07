package app

import (
	"net/http"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

// DeleteCategoriesHandler delete all categories entries effectively
// purging the entire tree.
func (a *App) DeleteCategoriesHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		contextLogger := log.WithContext(ctx)
		contextLogger.Info("app: DeleteCategoriesHandler started")

		// A catalog may only be purged if all catalog product associations are first purged.
		has, err := a.Service.HasProductCategoryRelations(ctx)
		if err != nil {
			contextLogger.Errorf("%+v", errors.Cause(err))
			w.WriteHeader(http.StatusInternalServerError) // 500 Internal Server Error
			return
		}
		if has {
			clientError(w, http.StatusConflict, ErrCodeAssocsExist, "OpDeleteCategories cannot be called whilst category to product associations exist")
			return
		}
		if err = a.Service.DeleteCategories(ctx); err != nil {
			contextLogger.Errorf("app: DeleteCategories(ctx) error: %v", errors.Cause(err))
			w.WriteHeader(http.StatusInternalServerError) // 500 Internal Server Error
			return
		}
		w.Header().Del("Content-Type")
		w.WriteHeader(http.StatusNoContent) // 204 No Content
	}
}
