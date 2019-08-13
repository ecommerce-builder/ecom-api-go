package app

import (
	"encoding/json"
	"fmt"
	"net/http"

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

		productID := chi.URLParam(r, "product_id")
		exists, err := a.Service.ProductExists(ctx, productID)
		if err != nil {
			contextLogger.Errorf("a.Service.ProductExists(ctx, productID=%q) failed: error %v", productID, err)
			w.WriteHeader(http.StatusInternalServerError) // 500 Internal Server Error
			return
		}
		if !exists {
			w.WriteHeader(http.StatusConflict) // 409 Conflict
			json.NewEncoder(w).Encode(struct {
				Status  int    `json:"status"`
				Code    string `json:"code"`
				Message string `json:"message"`
			}{
				http.StatusConflict,
				ErrCodeProductNotFound,
				fmt.Sprintf("product of productID=%q not found", productID),
			})
			return
		}

		imageRequestBody := imageRequestBody{}
		if err := json.NewDecoder(r.Body).Decode(&imageRequestBody); err != nil {
			http.Error(w, err.Error(), 400)
			return
		}
		exists, err = a.Service.ImagePathExists(ctx, imageRequestBody.Path)
		if err != nil {
			contextLogger.Errorf("service ImageExists(ctx, path=%q) error: %v", imageRequestBody.Path, err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		if exists {
			w.WriteHeader(http.StatusConflict)
			json.NewEncoder(w).Encode(struct {
				Code    int    `json:"code"`
				Message string `json:"message"`
			}{
				409,
				fmt.Sprintf("image with path=%s already exists", imageRequestBody.Path),
			})
			return
		}
		image, err := a.Service.CreateImageEntry(ctx, &productID, imageRequestBody.Path)
		if err != nil {
			contextLogger.Errorf("service CreateImageEntry(ctx, productID=%s, %s) error: %v", imageRequestBody.Path, productID, err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusCreated) // 201 Created
		json.NewEncoder(w).Encode(image)
	}
}
