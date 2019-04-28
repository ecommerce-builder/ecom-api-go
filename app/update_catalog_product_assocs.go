package app

import (
	"encoding/json"
	"fmt"
	"net/http"

	service "bitbucket.org/andyfusniakteam/ecom-api-go/service/firebase"
)

// UpdateCatalogProductAssocsHandler creates a handler to return the entire catalog
func (app *App) UpdateCatalogProductAssocsHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cpo := []*service.CatalogProductAssoc{}
		err := json.NewDecoder(r.Body).Decode(&cpo)
		if err != nil {
			http.Error(w, err.Error(), 400)
			return
		}

		for _, i := range cpo {
			fmt.Printf("%v\n", *i)
		}

		//app.Service.UpdateCatalogProductAssocs(r.Context(), cpo)
		w.Header().Del("Content-Type")
		w.WriteHeader(http.StatusNoContent) // 204 OK
	}
}
