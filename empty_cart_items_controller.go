package app

import (
	"fmt"
	"net/http"
	"os"

	"github.com/gorilla/mux"
)

// EmptyCartItemsController empties the cart of all items. This does not remove coupons.
func (a *App) EmptyCartItemsController() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		params := mux.Vars(r)

		err := a.Service.EmptyCartItems(params["ctid"])
		if err != nil {
			fmt.Fprintf(os.Stderr, "service EmptyCartItems(%s) error: %v", params["ctid"], err)
		}
		w.WriteHeader(http.StatusNoContent) // 204 No Content
	}
}
