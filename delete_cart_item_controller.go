package app

import (
	"net/http"

	"github.com/go-chi/chi"
)

// DeleteCartItemController handler
func (a *App) DeleteCartItemController() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctid := chi.URLParam(r, "ctid")
		sku := chi.URLParam(r, "sku")

		count, _ := a.Service.DeleteCartItem(r.Context(), ctid, sku)
		if count == 0 {
			w.WriteHeader(http.StatusNotFound) // 404 Not Found
			return
		}

		w.WriteHeader(http.StatusNoContent) // 204 No Content
	}
}
