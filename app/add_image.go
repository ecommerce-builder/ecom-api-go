package app

import (
	"encoding/json"
	"net/http"

	service "bitbucket.org/andyfusniakteam/ecom-api-go/service/firebase"
	log "github.com/sirupsen/logrus"
)

type addImageRequestBody struct {
	ProductID *string `json:"product_id"`
	Path      *string `json:"path"`
}

func validateAddImageRequest(request *addImageRequestBody) (bool, string) {
	// product_id attribute
	if request.ProductID == nil {
		return false, "product_id attribute must be set"
	}
	if !IsValidUUID(*request.ProductID) {
		return false, "product_id attribute must be a valid v4 uuid"
	}

	// path attribute
	if request.Path == nil {
		return false, "path attribute must be set"
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
		dec := json.NewDecoder(r.Body)
		dec.DisallowUnknownFields()
		if err := dec.Decode(&request); err != nil {
			clientError(w, http.StatusBadRequest, ErrCodeBadRequest, err.Error()) // 400
			return
		}

		valid, message := validateAddImageRequest(&request)
		if !valid {
			clientError(w, http.StatusBadRequest, ErrCodeBadRequest, message) // 400
			return
		}

		image, err := a.Service.CreateImage(ctx, *request.ProductID, *request.Path)
		if err == service.ErrProductNotFound {
			clientError(w, http.StatusNotFound, ErrCodeProductNotFound, "product not found") // 404
			return
		}
		if err != nil {
			contextLogger.Errorf("app: a.Service.CreateImage(ctx, productID=%q, path=%q) error: %+v", *request.ProductID, *request.Path, err)
			w.WriteHeader(http.StatusInternalServerError) // 500
			return
		}
		w.WriteHeader(http.StatusCreated) // 201 Created
		json.NewEncoder(w).Encode(&image)
	}
}
