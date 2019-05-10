package app

import (
	"encoding/json"
	"net/http"

	"bitbucket.org/andyfusniakteam/ecom-api-go/utils/nestedset"
)

// UpdateCatalogHandler creates an HTTP handler that updates the catalog.
func (a *App) UpdateCatalogHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		catalog := nestedset.Node{}
		err := json.NewDecoder(r.Body).Decode(&catalog)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest) // 400 Bad Request
			json.NewEncoder(w).Encode(struct {
				Code    int    `json:"code"`
				Message string `json:"message"`
			}{
				400,
				err.Error(),
			})
			return
		}
		defer r.Body.Close()
		a.Service.UpdateCatalog(r.Context(), &catalog)
		w.Header().Del("Content-Type")
		w.WriteHeader(http.StatusNoContent) // 204 No Content
	}
}
