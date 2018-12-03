package app

import (
	"fmt"
	"net/http"
	"os"

	"github.com/go-chi/chi"
)

// EmptyCartItemsController empties the cart of all items. This does not remove coupons.
func (a *App) EmptyCartItemsController() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctid := chi.URLParam(r, "ctid")

		err := a.Service.EmptyCartItems(ctid)
		if err != nil {
			fmt.Fprintf(os.Stderr, "service EmptyCartItems(%s) error: %v", ctid, err)
		}
		w.WriteHeader(http.StatusNoContent) // 204 No Content
	}
}
