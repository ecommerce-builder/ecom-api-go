package app

import (
	"net/http"

	service "bitbucket.org/andyfusniakteam/ecom-api-go/service/firebase"
	"github.com/go-chi/chi"
	log "github.com/sirupsen/logrus"
)

// DeleteImageHandler create a handler that deletes an image for the
// product with the given SKU.
func (a *App) DeleteImageHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		contextLogger := log.WithContext(ctx)
		contextLogger.Info("app: DeleteImageHandler started")

		imageID := chi.URLParam(r, "id")
		err := a.Service.DeleteImage(ctx, imageID)
		if err == service.ErrImageNotFound {
			clientError(w, http.StatusNotFound, ErrCodeImageNotFound, "image not found")
			return
		}
		if err != nil {
			contextLogger.Errorf("app: DeleteImage(ctx, imageID=%q) failed: %+v", imageID, err)
			w.WriteHeader(http.StatusInternalServerError) // 500 Internal Server Error
			return
		}
		w.Header().Del("Content-Type")
		w.WriteHeader(http.StatusNoContent) // 204 No Content
	}
}
