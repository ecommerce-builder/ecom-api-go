package app

import (
	"fmt"
	"net/http"
	"os"
)

// DeleteCatalogProductAssocsHandler deletes all catalog product associations.
func (a *App) DeleteCatalogProductAssocsHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		_, err := a.Service.DeleteCatalogProductAssocs(r.Context())
		if err != nil {
			fmt.Fprintf(os.Stderr, "service DeleteCatalogProductAssocs(ctx) error: %v", err)
			return
		}
		w.Header().Del("Content-Type")
		w.WriteHeader(http.StatusOK) // 200 OK
	}
}
