package app

import (
	"database/sql"
	"fmt"
	"net/http"
	"os"

	"github.com/go-chi/chi"
)

// DeleteImageHandler create a handler that deletes an image for the
// product with the given SKU.
func (a *App) DeleteImageHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		uuid := chi.URLParam(r, "uuid")
		exists, err := a.Service.ImageUUIDExists(r.Context(), uuid)
		if err != nil {
			fmt.Fprintf(os.Stderr, "ImageUUIDExists(ctx, %q) failed: %v", uuid, err)
			w.WriteHeader(http.StatusInternalServerError) // 500 Internal Server Error
			return
		}
		if !exists {
			w.WriteHeader(http.StatusNotFound) // 404 Not Found
			return
		}
		if err := a.Service.DeleteImage(r.Context(), uuid); err != nil {
			if err == sql.ErrNoRows {
				w.WriteHeader(http.StatusNotFound)
				return
			}
			fmt.Fprintf(os.Stderr, "delete image uuid=%q failed: %v", uuid, err)
			w.WriteHeader(http.StatusInternalServerError) // 500 Internal Server Error
			return
		}
		w.Header().Del("Content-Type")
		w.WriteHeader(http.StatusNoContent) // 204 No Content
	}
}
