package app

import (
	"encoding/json"
	"net/http"

	"bitbucket.org/andyfusniakteam/ecom-api-go/model/postgres"
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

		sku := chi.URLParam(r, "sku")
		product, err := a.Service.GetProduct(ctx, sku)
		if err != nil {
			if err == postgres.ErrProductNotFound {
				w.WriteHeader(http.StatusNotFound)
				return
			}
			contextLogger.Errorf("app: GetProduct(ctx, %q) error: %+v", sku, err)
			w.WriteHeader(http.StatusInternalServerError) // 500 Internal Server Error
			return
		}
		w.WriteHeader(http.StatusOK) // 200 OK
		json.NewEncoder(w).Encode(*product)
	}
}
