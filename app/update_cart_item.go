package app

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/go-chi/chi"
)

// UpdateCartItemHandler creates a handler to add an item to a given cart
func (app *App) UpdateCartItemHandler() http.HandlerFunc {
	type qtyRequestBody struct {
		Qty int `json:"qty"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		uuid := chi.URLParam(r, "uuid")
		sku := chi.URLParam(r, "sku")
		o := qtyRequestBody{}
		err := json.NewDecoder(r.Body).Decode(&o)
		if err != nil {
			http.Error(w, err.Error(), 400)
			return
		}
		cart, err := app.Service.UpdateCartItem(r.Context(), uuid, sku, o.Qty)
		if err != nil {
			fmt.Fprintf(os.Stderr, "service UpdateCartItem(ctx, %s, %s, %d) error: %v", uuid, sku, o.Qty, err)
			w.WriteHeader(http.StatusInternalServerError) // 500 Internal Server Error
			return
		}
		w.WriteHeader(http.StatusOK) // 200 OK
		json.NewEncoder(w).Encode(*cart)
	}
}
