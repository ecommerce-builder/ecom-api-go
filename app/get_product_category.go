package app

import (
	"encoding/json"
	"net/http"

	service "bitbucket.org/andyfusniakteam/ecom-api-go/service/firebase"
	"github.com/go-chi/chi"
	log "github.com/sirupsen/logrus"
)

// GetProductCategoryHandler creates a handler function that returns a
// price list.
func (a *App) GetProductCategoryHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		contextLogger := log.WithContext(ctx)
		contextLogger.Info("App: GetProductCategoryHandler called")

		productCategoryID := chi.URLParam(r, "id")
		productCategory, err := a.Service.GetProductCategory(ctx, productCategoryID)
		if err != nil {
			if err == service.ErrProductCategoryNotFound {
				json.NewEncoder(w).Encode(struct {
					Status  int    `json:"status"`
					Code    string `json:"code"`
					Message string `json:"message"`
				}{
					http.StatusNotFound,
					ErrCodeProductCategoryNotFound,
					"product to category association not found",
				})
				return
			}
			contextLogger.Errorf("app: a.Service.GetProductCategory(ctx, productCategoryID=%q)", productCategoryID)
			w.WriteHeader(http.StatusInternalServerError) // 500 Internal Server Error
		}
		w.WriteHeader(http.StatusOK) // 200 OK
		json.NewEncoder(w).Encode(&productCategory)
	}
}
