package app

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/go-chi/chi"
)

// AddImageHandler creates a handler to add an images to a product.
func (a *App) AddImageHandler() http.HandlerFunc {
	type imageRequestBody struct {
		Path string `json:"path"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		sku := chi.URLParam(r, "sku")
		exists, err := a.Service.ProductExists(ctx, sku)
		if err != nil {
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
				ErrCodeProductSKUNotFound,
				fmt.Sprintf("product with sku=%s not found", sku),
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
			fmt.Fprintf(os.Stderr, "service ImageExists(ctx, %s, %s) error: %v", imageRequestBody.Path, sku, err)
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
		image, err := a.Service.CreateImageEntry(ctx, sku, imageRequestBody.Path)
		if err != nil {
			fmt.Fprintf(os.Stderr, "service CreateImageEntry(ctx, %s, %s) error: %v", imageRequestBody.Path, sku, err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusCreated) // 201 Created
		json.NewEncoder(w).Encode(image)
	}
}
