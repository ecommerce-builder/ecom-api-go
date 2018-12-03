package app

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/go-chi/chi"
)

// GetCartItemsController handler
func (a *App) GetCartItemsController() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctid := chi.URLParam(r, "ctid")

		cartItems, err := a.Service.GetCartItems(ctid)
		if err != nil {
			fmt.Fprintf(os.Stderr, "service GetCartItems(%s) error: %v", ctid, err)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK) // 200 OK
		json.NewEncoder(w).Encode(cartItems)
	}
}
