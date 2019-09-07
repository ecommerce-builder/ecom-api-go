package app

import (
	"encoding/json"
	"net/http"

	service "bitbucket.org/andyfusniakteam/ecom-api-go/service/firebase"
	log "github.com/sirupsen/logrus"
)

type addImageRequestBody struct {
	ProductID string `json:"product_id"`
	Path      string `json:"path"`
}

func validateAddImageRequest(request *addImageRequestBody) (bool, string) {
	if request.ProductID == "" {
		return false, "product_id attribute is required"
	}
	if request.Path == "" {
		return false, "path attribute is required"
	}
	return true, ""
}

// AddImageHandler creates a handler to add an images to a product.
func (a *App) AddImageHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		contextLogger := log.WithContext(ctx)
		contextLogger.Info("app: AddImageHandler started")

		request := addImageRequestBody{}
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			clientError(w, http.StatusBadRequest, ErrCodeBadRequest, err.Error())
			return
		}

		valid, message := validateAddImageRequest(&request)
		if !valid {
			clientError(w, http.StatusBadRequest, ErrCodeBadRequest, message)
			return
		}

		image, err := a.Service.CreateImage(ctx, request.ProductID, request.Path)
		if err != nil {
			if err == service.ErrProductNotFound {
				clientError(w, http.StatusNotFound, ErrCodeProductNotFound, "product not found")
				return
			}
			contextLogger.Errorf("app: CreateImageEntry(ctx, productID=%q, %s) error: %+v", request.Path, request.ProductID, err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusCreated) // 201 Created
		json.NewEncoder(w).Encode(image)
	}
}
