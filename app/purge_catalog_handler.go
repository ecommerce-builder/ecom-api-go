package app

import (
	"fmt"
	"net/http"
	"os"
)

// PurgeCatalogHandler purges the catalog hierarchy.
func (a *App) PurgeCatalogHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		err := a.Service.DeleteCatalog(r.Context())
		if err != nil {
			fmt.Fprintf(os.Stderr, "service DeleteCatalog(ctx) error: %v", err)
			return
		}
		w.Header().Del("Content-Type")
		w.WriteHeader(http.StatusNoContent) // 204 No Content
	}
}
