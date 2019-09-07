package app

import (
	"encoding/json"
	"net/http"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

// GetCategoriesTreeHandler creates a handler to return the entire categories
// represented as a tree structure.
func (app *App) GetCategoriesTreeHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		contextLogger := log.WithContext(ctx)
		contextLogger.Info("app: GetCategoriesHandler called")

		tree, err := app.Service.GetCategoriesTree(ctx)
		if err != nil {
			contextLogger.Errorf("app: service GetCatalog(ctx) error: %+v", errors.WithStack(err))
			w.WriteHeader(http.StatusInternalServerError) // 500
			return
		}
		if tree == nil {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("{}"))
			return
		}
		if err != nil {
			contextLogger.Errorf("app: GetCatalog(ctx) error: %+v", errors.WithStack(err))
			w.WriteHeader(http.StatusInternalServerError) // 500
			return
		}
		w.WriteHeader(http.StatusOK) // 200 OK
		if err = json.NewEncoder(w).Encode(tree); err != nil {
			contextLogger.Errorf("json encode failed with error: %+v", err)
			w.WriteHeader(http.StatusInternalServerError) // 500 Internal Server Error
			return
		}
	}
}
