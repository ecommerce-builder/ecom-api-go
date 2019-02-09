package app

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/go-chi/chi"
)

// UpdateCartItemHandler creates a handler to add an item to a given cart
func (a *App) UpdateCartItemHandler() http.HandlerFunc {
	type qtyRequestBody struct {
		Qty int `json:"qty"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		ctid := chi.URLParam(r, "ctid")
		sku := chi.URLParam(r, "sku")

		o := qtyRequestBody{}
		err := json.NewDecoder(r.Body).Decode(&o)
		if err != nil {
			http.Error(w, err.Error(), 400)
			return
		}

		cart, err := a.Service.UpdateCartItem(r.Context(), ctid, sku, o.Qty)
		if err != nil {
			fmt.Fprintf(os.Stderr, "service UpdateCartItem(ctx, %s, %s, %d) error: %v", ctid, sku, o.Qty, err)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK) // 200 OK
		json.NewEncoder(w).Encode(*cart)
	}
}
