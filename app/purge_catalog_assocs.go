package app

import (
	"fmt"
	"net/http"
	"os"
)

// PurgeCatalogAssocsHandler deletes all catalog product associations.
func (a *App) PurgeCatalogAssocsHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		_, err := a.Service.DeleteCategoryAssocs(r.Context())
		if err != nil {
			fmt.Fprintf(os.Stderr, "service DeleteCategoryAssocs(ctx) error: %+v", err)
			w.WriteHeader(http.StatusInternalServerError) // 500 Internal Server Error
			return
		}
		w.Header().Del("Content-Type")
		w.WriteHeader(http.StatusNoContent) // 204 No Content
	}
}