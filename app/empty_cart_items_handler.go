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
		ctid := chi.URLParam(r, "ctid")

		err := a.Service.EmptyCartItems(r.Context(), ctid)
		if err != nil {
			fmt.Fprintf(os.Stderr, "service EmptyCartItems(ctx, %s) error: %v", ctid, err)
		}
		w.Header().Del("Content-Type")
		w.WriteHeader(http.StatusNoContent) // 204 No Content
	}
}
