package app

import (
	"encoding/json"
	"net/http"

	service "bitbucket.org/andyfusniakteam/ecom-api-go/service/firebase"
	log "github.com/sirupsen/logrus"

	"github.com/go-chi/chi"
)

// AddImageHandler creates a handler to add an images to a product.
func (a *App) AddImageHandler() http.HandlerFunc {
	type imageRequestBody struct {
		Path string `json:"path"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		contextLogger := log.WithContext(ctx)
		contextLogger.Info("App: AddImageHandler started")

		productID := chi.URLParam(r, "id")

		imageRequestBody := imageRequestBody{}
		if err := json.NewDecoder(r.Body).Decode(&imageRequestBody); err != nil {
			http.Error(w, err.Error(), 400)
			return
		}

		image, err := a.Service.CreateImage(ctx, productID, imageRequestBody.Path)
		if err != nil {
			if err == service.ErrProductNotFound {
				w.WriteHeader(http.StatusNotFound) // Not Found
				json.NewEncoder(w).Encode(struct {
					Status  int    `json:"status"`
					Code    string `json:"code"`
					Message string `json:"message"`
				}{
					http.StatusNotFound,
					ErrCodeProductNotFound,
					"product not found",
				})
				return
			}
			contextLogger.Errorf("service CreateImageEntry(ctx, productID=%s, %s) error: %v", imageRequestBody.Path, productID, err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusCreated) // 201 Created
		json.NewEncoder(w).Encode(image)
	}
}
