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
		uuid := chi.URLParam(r, "uuid")
		cartItems, err := a.Service.GetCartItems(r.Context(), uuid)
		if err != nil {
			fmt.Fprintf(os.Stderr, "service GetCartItems(%s) error: %v", uuid, err)
			w.WriteHeader(http.StatusInternalServerError) // 500 Internal Server Error
			return
		}
		w.WriteHeader(http.StatusOK) // 200 OK
		json.NewEncoder(w).Encode(cartItems)
	}
}
