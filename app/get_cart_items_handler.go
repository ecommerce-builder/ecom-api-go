package app

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/go-chi/chi"
)

// GetCartItemsHandler returns a list of all cart items
func (a *App) GetCartItemsHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctid := chi.URLParam(r, "ctid")

		cartItems, err := a.Service.GetCartItems(r.Context(), ctid)
		if err != nil {
			fmt.Fprintf(os.Stderr, "service GetCartItems(%s) error: %v", ctid, err)
			return
		}
		w.WriteHeader(http.StatusOK) // 200 OK
		json.NewEncoder(w).Encode(cartItems)
	}
}
