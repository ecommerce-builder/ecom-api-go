package app

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/pkg/errors"
)

// PurgeCatalogHandler purges the catalog hierarchy.
func (a *App) PurgeCatalogHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		// A catalog may only be purged if all catalog product associations are first purged.
		has, err := a.Service.HasCatalogProductAssocs(ctx)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%+v", errors.Cause(err))
			w.WriteHeader(http.StatusInternalServerError) // 500 Internal Server Error
			return
		}
		if has {
			w.WriteHeader(http.StatusConflict) // 409 Conflict
			json.NewEncoder(w).Encode(struct {
				Code    int    `json:"code"`
				Message string `json:"message"`
			}{
				http.StatusConflict,
				"Catalog cannot be purged before catalog product associations have first been purged.",
			})
			return
		}

		err = a.Service.DeleteCatalog(ctx)
		if err != nil {
			fmt.Fprintf(os.Stderr, "service DeleteCatalog(ctx) error: %v", errors.Cause(err))
			w.WriteHeader(http.StatusInternalServerError) // 500 Internal Server Error
			return
		}
		w.Header().Del("Content-Type")
		w.WriteHeader(http.StatusNoContent) // 204 No Content
	}
}
