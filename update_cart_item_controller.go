package app

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/gorilla/mux"
)

// UpdateCartItemController creates a hanlder that adds an item to a given cart
func (a *App) UpdateCartItemController() http.HandlerFunc {
	type qtyRequestBody struct {
		Qty int `json:"qty"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		params := mux.Vars(r)

		o := qtyRequestBody{}
		err := json.NewDecoder(r.Body).Decode(&o)
		if err != nil {
			http.Error(w, err.Error(), 400)
			return
		}

		cart, err := a.Service.UpdateCartItem(params["ctid"], params["sku"], o.Qty)
		if err != nil {
			fmt.Fprintf(os.Stderr, "service UpdateCartItem(%s, %s, %d) error: %v", params["ctid"], params["sku"], o.Qty, err)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK) // 200 OK
		json.NewEncoder(w).Encode(*cart)
	}
}
