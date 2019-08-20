package app

import (
	"database/sql"
	"net/http"

	"github.com/go-chi/chi"
	log "github.com/sirupsen/logrus"
)

// DeleteImageHandler create a handler that deletes an image for the
// product with the given SKU.
func (a *App) DeleteImageHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		contextLogger := log.WithContext(ctx)
		contextLogger.Info("App: DeleteImageHandler started")

		imageID := chi.URLParam(r, "id")
		exists, err := a.Service.ImageUUIDExists(ctx, imageID)
		if err != nil {
			contextLogger.Errorf("ImageUUIDExists(ctx, imageID=%q) failed: %v", imageID, err)
			w.WriteHeader(http.StatusInternalServerError) // 500 Internal Server Error
			return
		}
		if !exists {
			w.WriteHeader(http.StatusNotFound) // 404 Not Found
			return
		}
		if err := a.Service.DeleteImage(ctx, imageID); err != nil {
			if err == sql.ErrNoRows {
				w.WriteHeader(http.StatusNotFound)
				return
			}
			contextLogger.Errorf("delete image imageID=%q failed: %v", imageID, err)
			w.WriteHeader(http.StatusInternalServerError) // 500 Internal Server Error
			return
		}
		w.Header().Del("Content-Type")
		w.WriteHeader(http.StatusNoContent) // 204 No Content
	}
}
