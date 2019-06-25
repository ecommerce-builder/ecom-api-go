package app

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

// GetCatalogHandler creates a handler to return the entire catalog
func (app *App) GetCatalogHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		log.WithContext(ctx).Info("App: GetCatalogHandler called")
		tree, err := app.Service.GetCatalog(ctx)
		if tree == nil {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("{}"))
			return
		}
		if err != nil {
			fmt.Fprintf(os.Stderr, "service GetCatalog(ctx) error: %+v", errors.WithStack(err))
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
		if err = json.NewEncoder(w).Encode(tree); err != nil {
			fmt.Fprintf(os.Stderr, "%+v", err)
			w.WriteHeader(http.StatusInternalServerError) // 500 Internal Server Error
			return
		}
	}
}
