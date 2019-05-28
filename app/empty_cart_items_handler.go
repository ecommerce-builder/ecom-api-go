package app

import (
	"fmt"
	"net/http"
	"os"

	"github.com/go-chi/chi"
)

// EmptyCartItemsHandler empties the cart of all items. This does not remove coupons.
func (a *App) EmptyCartItemsHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		uuid := chi.URLParam(r, "uuid")
		if err := a.Service.EmptyCartItems(r.Context(), uuid); err != nil {
			fmt.Fprintf(os.Stderr, "service EmptyCartItems(ctx, %s) error: %v", uuid, err)
			w.WriteHeader(http.StatusInternalServerError) // 500 Internal Server Error
			return
		}
		w.Header().Del("Content-Type")
		w.WriteHeader(http.StatusNoContent) // 204 No Content
	}
}
