package app

import (
	"encoding/json"
	"net/http"

	service "bitbucket.org/andyfusniakteam/ecom-api-go/service/firebase"
	"github.com/go-chi/chi"
	log "github.com/sirupsen/logrus"
)

// GetImageHandler returns a handler function that gets a single image by UUID.
func (a *App) GetImageHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		contextLogger := log.WithContext(ctx)
		contextLogger.Info("app: GetImageHandler called")

		imageID := chi.URLParam(r, "id")
		image, err := a.Service.GetImage(ctx, imageID)
		if err == service.ErrImageNotFound {
			clientError(w, http.StatusNotFound, ErrCodeImageNotFound, "image not found")
			return
		}
		if err != nil {
			contextLogger.Errorf("app: GetImage(ctx, imageID=%q) error: %+v", imageID, err)
			w.WriteHeader(http.StatusInternalServerError) // 500 Internal Server Error
			return
		}
		w.WriteHeader(http.StatusOK) // 200 OK
		json.NewEncoder(w).Encode(image)
	}
}
