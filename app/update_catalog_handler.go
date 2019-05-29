package app

import (
	"encoding/json"
	"net/http"

	"bitbucket.org/andyfusniakteam/ecom-api-go/service/firebase"
)

// UpdateCatalogHandler creates an HTTP handler that updates the catalog.
func (a *App) UpdateCatalogHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cats := firebase.Category{}
		if err := json.NewDecoder(r.Body).Decode(&cats); err != nil {
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
		if err := a.Service.UpdateCatalog(r.Context(), &cats); err != nil {
			w.WriteHeader(http.StatusInternalServerError) // 500 Internal Server Error
			return
		}
		w.Header().Del("Content-Type")
		w.WriteHeader(http.StatusNoContent) // 204 No Content
	}
}
