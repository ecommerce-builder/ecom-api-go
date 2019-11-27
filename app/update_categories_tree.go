package app

import (
	"encoding/json"
	"fmt"
	"net/http"

	service "bitbucket.org/andyfusniakteam/ecom-api-go/service/firebase"
	log "github.com/sirupsen/logrus"
)

// UpdateCategoriesTreeHandler creates an HTTP handler that updates all the categories
// using a tree structure.
func (a *App) UpdateCategoriesTreeHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		contextLogger := log.WithContext(ctx)
		contextLogger.Info("app: UpdateCategoriesTreeHandler started")

		catRequest := service.CategoryRequest{}
		dec := json.NewDecoder(r.Body)
		dec.DisallowUnknownFields()
		if err := dec.Decode(&catRequest); err != nil {
			clientError(w, http.StatusBadRequest, ErrCodeBadRequest,
				err.Error()) // 400
			return
		}
		defer r.Body.Close()

		err := a.Service.UpdateCatalog(ctx, &catRequest)
		if err == service.ErrCategoriesInUse {
			clientError(w, http.StatusConflict, ErrCodeCategoriesInUse,
				"one or more categories in use - check promo rules") // 409
			return
		}
		if err == service.ErrAssocsAlreadyExist {
			clientError(w, http.StatusConflict, ErrCodeAssocsExist,
				fmt.Sprintf("product to category relations already exist")) // 409
			return
		}
		if err != nil {
			contextLogger.Errorf("app: UpdateCatalog(ctx, cats) failed: %+v", err)
			w.WriteHeader(http.StatusInternalServerError) // 500
			return
		}
		w.Header().Del("Content-Type")
		w.Header().Set("Content-Length", "0")
		w.WriteHeader(http.StatusNoContent) // 204
	}
}
