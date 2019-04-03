package app

import (
	"net/http"

	"github.com/go-chi/chi"
)

// DeleteCartItemHandler creates a handler to delete an item from the cart with the given cart UUID.
func (a *App) DeleteCartItemHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctid := chi.URLParam(r, "ctid")
		sku := chi.URLParam(r, "sku")

		count, _ := a.Service.DeleteCartItem(r.Context(), ctid, sku)
		if count == 0 {
			w.WriteHeader(http.StatusNotFound) // 404 Not Found
			return
		}
		w.Header().Del("Content-Type")
		w.WriteHeader(http.StatusNoContent) // 204 No Content
	}
}
