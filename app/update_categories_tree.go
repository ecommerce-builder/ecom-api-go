package app

import (
	"encoding/json"
	"fmt"
	"net/http"

	"bitbucket.org/andyfusniakteam/ecom-api-go/service/firebase"
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
			clientError(w, http.StatusBadRequest, ErrCodeBadRequest, err.Error())
			return
		}
		defer r.Body.Close()
		err := a.Service.UpdateCatalog(ctx, &catRequest)
		if err == firebase.ErrAssocsAlreadyExist {
			clientError(w, http.StatusConflict, ErrCodeAssocsExist, fmt.Sprintf("product to category relations already exist"))
			return
		}
		if err != nil {
			contextLogger.Errorf("app: UpdateCatalog(ctx, cats) failed: %+v", err)
			w.WriteHeader(http.StatusInternalServerError) // 500 Internal Server Error
			return
		}
		w.Header().Del("Content-Type")
		w.WriteHeader(http.StatusNoContent) // 204 No Content
	}
}
