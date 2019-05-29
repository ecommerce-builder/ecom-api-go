package app

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/go-chi/chi"
)

// AddItemToCartHandler creates a handler to add an item to a given cart
func (a *App) AddItemToCartHandler() http.HandlerFunc {
	type itemRequestBody struct {
		SKU string `json:"sku"`
		Qty int    `json:"qty"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		uuid := chi.URLParam(r, "uuid")
		o := itemRequestBody{}
		if err := json.NewDecoder(r.Body).Decode(&o); err != nil {
			http.Error(w, err.Error(), 400)
			return
		}
		item, err := a.Service.AddItemToCart(r.Context(), uuid, o.SKU, o.Qty)
		if err != nil {
			fmt.Fprintf(os.Stderr, "service AddItemToCart(%s, %s, %d) error: %v", uuid, o.SKU, o.Qty, err)
			w.WriteHeader(http.StatusInternalServerError) // 500 Internal Server Error
			return
		}
		w.WriteHeader(http.StatusCreated) // 201 Created
		json.NewEncoder(w).Encode(item)
	}
}
