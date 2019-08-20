package app

import (
	"encoding/json"
	"net/http"

	service "bitbucket.org/andyfusniakteam/ecom-api-go/service/firebase"
	"github.com/go-chi/chi"
	log "github.com/sirupsen/logrus"
)

// GetProductHandler returns a handler function that gets a product by SKU
// containing product and image data.
func (a *App) GetProductHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		contextLogger := log.WithContext(ctx)
		contextLogger.Info("App: GetProductHandler called")

		productID := chi.URLParam(r, "id")
		product, err := a.Service.GetProduct(ctx, productID)
		if err != nil {
			if err == service.ErrProductNotFound {
				w.WriteHeader(http.StatusNotFound)
				return
			}
			contextLogger.Errorf("app: GetProduct(ctx, productID=%q) error: %+v", productID, err)
			w.WriteHeader(http.StatusInternalServerError) // 500 Internal Server Error
			return
		}
		w.WriteHeader(http.StatusOK) // 200 OK
		json.NewEncoder(w).Encode(*product)
	}
}
