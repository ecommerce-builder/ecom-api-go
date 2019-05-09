package app

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/pkg/errors"
)

// GetCatalogHandler creates a handler to return the entire catalog
func (app *App) GetCatalogHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tree, err := app.Service.GetCatalog(r.Context())
		if err != nil {
			fmt.Fprintf(os.Stderr, "service GetCatalog(ctx) error: %+v", errors.WithStack(err))
			w.WriteHeader(http.StatusInternalServerError) // 500
			json.NewEncoder(w).Encode(struct {
				Code    int    `json:"code"`
				Message string `json:"message"`
			}{
				500,
				err.Error(),
			})
			return
		}
		w.WriteHeader(http.StatusOK) // 200 OK
		err = json.NewEncoder(w).Encode(tree)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%+v", err)
		}
	}
}
