package app

import (
	"encoding/json"
	"net/http"

	service "bitbucket.org/andyfusniakteam/ecom-api-go/service/firebase"
	log "github.com/sirupsen/logrus"
)

// ListProductImagesHandler creates a handler function that returns a
// slice of ProductImages for the product with the given SKU.
func (a *App) ListProductImagesHandler() http.HandlerFunc {
	type listResponse struct {
		Object string           `json:"object"`
		Data   []*service.Image `json:"data"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		contextLogger := log.WithContext(ctx)
		contextLogger.Info("App: ListProductImagesHandler started")

		productID := r.URL.Query().Get("product_id")
		if productID == "" {
			w.WriteHeader(http.StatusBadRequest) // 400 Bad Request
			json.NewEncoder(w).Encode(struct {
				Status  int    `json:"status"`
				Code    string `json:"code"`
				Message string `json:"message"`
			}{
				http.StatusBadRequest,
				ErrCodeBadRequest,
				"query parameter product_id must be set",
			})
			return
		}

		products, err := a.Service.GetImagesByProductID(ctx, productID)
		if err != nil {
			contextLogger.Errorf("service ListProductImages(ctx, productID=%q) error: %+v", productID, err)
			w.WriteHeader(http.StatusInternalServerError) // 500 Internal Server Error
			return
		}

		list := listResponse{
			Object: "list",
			Data:   products,
		}
		w.WriteHeader(http.StatusOK) // 200 OK
		json.NewEncoder(w).Encode(&list)
	}
}
