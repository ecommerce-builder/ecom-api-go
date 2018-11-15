package app

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/gorilla/mux"
)

// AddItemToCartController creates a hanlder that adds an item to a given cart
func (a *App) AddItemToCartController() http.HandlerFunc {
	type itemRequestBody struct {
		Sku string `json:"sku"`
		Qty int    `json:"qty"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		params := mux.Vars(r)

		o := itemRequestBody{}
		err := json.NewDecoder(r.Body).Decode(&o)
		if err != nil {
			http.Error(w, err.Error(), 400)
			return
		}

		item, err := a.Service.AddItemToCart(params["ctid"], o.Sku, o.Qty)
		if err != nil {
			fmt.Fprintf(os.Stderr, "service AddItemToCart(%s, %s, %d) error: %v", params["ctid"], o.Sku, o.Qty, err)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated) // 201 Created
		json.NewEncoder(w).Encode(item)
	}
}
