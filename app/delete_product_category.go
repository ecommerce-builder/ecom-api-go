package app

import (
	"encoding/json"
	"net/http"

	service "bitbucket.org/andyfusniakteam/ecom-api-go/service/firebase"
	"github.com/go-chi/chi"
	log "github.com/sirupsen/logrus"
)

// DeleteProductCategoryHandler create a handler to delete a product resource.
func (a *App) DeleteProductCategoryHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		contextLogger := log.WithContext(ctx)
		contextLogger.Info("App: DeleteProductCategoryHandler started")

		productCategoryID := chi.URLParam(r, "id")
		if err := a.Service.DeleteProductCategory(ctx, productCategoryID); err != nil {
			if err == service.ErrProductCategoryNotFound {
				w.WriteHeader(http.StatusNotFound) // 404 Not Found
				json.NewEncoder(w).Encode(struct {
					Status  int    `json:"status"`
					Code    string `json:"code"`
					Message string `json:"message"`
				}{
					http.StatusNotFound,
					ErrCodeProductCategoryNotFound,
					"product not found",
				})
				return
			}
			contextLogger.Errorf("app: a.Service.DeleteProductCategory(ctx, productCategoryID=%q) failed: %v", productCategoryID, err)
			w.WriteHeader(http.StatusInternalServerError) // 500 Internal Server Error
			return
		}
		w.Header().Del("Content-Type")
		w.Header().Set("Content-Length", "0")
		w.WriteHeader(http.StatusNoContent) // 204 No Content
	}
}
