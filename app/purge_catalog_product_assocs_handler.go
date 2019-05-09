package app

import (
	"fmt"
	"net/http"
	"os"
)

// PurgeCatalogProductAssocsHandler deletes all catalog product associations.
func (a *App) PurgeCatalogProductAssocsHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		_, err := a.Service.DeleteCatalogProductAssocs(r.Context())
		if err != nil {
			fmt.Fprintf(os.Stderr, "service DeleteCatalogProductAssocs(ctx) error: %+v", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.Header().Del("Content-Type")
		w.WriteHeader(http.StatusNoContent) // 204 No Content
	}
}
