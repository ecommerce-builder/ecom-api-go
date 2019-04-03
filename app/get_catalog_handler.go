package app

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
)

// GetCatalogHandler creates a handler to return the entire catalog
func (app *App) GetCatalogHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		catalog, err := app.Service.GetCatalog(r.Context())
		if err != nil {
			fmt.Fprintf(os.Stderr, "service GetCatalog(ctx) error: %v", err)
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
		json.NewEncoder(w).Encode(catalog)
	}
}
