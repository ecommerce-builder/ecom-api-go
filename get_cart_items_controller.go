package app

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/gorilla/mux"
)

// GetCartItemsController handler
func (a *App) GetCartItemsController() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		params := mux.Vars(r)

		cartItems, err := a.Service.GetCartItems(params["ctid"])
		if err != nil {
			fmt.Fprintf(os.Stderr, "service GetCartItems(%s) error: %v", params["ctid"], err)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK) // 200 OK
		json.NewEncoder(w).Encode(cartItems)
	}
}
