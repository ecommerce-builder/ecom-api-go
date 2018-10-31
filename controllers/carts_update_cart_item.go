package controllers

import (
	"encoding/json"
	"net/http"

	"bitbucket.org/andyfusniakteam/ecom-api-go/services"
	"github.com/gorilla/mux"
)

type qtyRequestBody struct {
	Qty int `json:"qty"`
}

// UpdateCartItem handler
func UpdateCartItem(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)

	o := qtyRequestBody{}
	err := json.NewDecoder(r.Body).Decode(&o)
	if err != nil {
		http.Error(w, err.Error(), 400)
		return
	}

	cart, _ := services.UpdateCartItem(params["ctid"], params["sku"], o.Qty)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK) // 200 OK
	json.NewEncoder(w).Encode(cart)
}
