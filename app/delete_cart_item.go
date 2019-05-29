package app

import (
	"fmt"
	"net/http"
	"os"

	"github.com/go-chi/chi"
)

// DeleteCartItemHandler creates a handler to delete an item from the cart with the given cart UUID.
func (a *App) DeleteCartItemHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		uuid := chi.URLParam(r, "uuid")
		sku := chi.URLParam(r, "sku")
		count, err := a.Service.DeleteCartItem(r.Context(), uuid, sku)
		if count == 0 {
			w.WriteHeader(http.StatusNotFound) // 404 Not Found
			return
		}
		if err != nil {
			fmt.Fprintf(os.Stderr, "DeleteCartItem(ctx, %q, %q) failed error: %v", uuid, sku, err)
			w.WriteHeader(http.StatusInternalServerError) // 500 Internal Server Error
			return
		}
		w.Header().Del("Content-Type")
		w.WriteHeader(http.StatusNoContent) // 204 No Content
	}
}
