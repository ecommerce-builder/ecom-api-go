package app

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
)

// GetCatalogProductAssocsHandler creates a handler to return the entire catalog
func (app *App) GetCatalogProductAssocsHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cpo, err := app.Service.GetCatalogProductAssocs(r.Context())
		if err != nil {
			fmt.Fprintf(os.Stderr, "service GetCatalogProductAssocs(ctx) error: %v", err)
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
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK) // 200 OK
		json.NewEncoder(w).Encode(cpo)
	}
}
