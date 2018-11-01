package controllers

import (
	"encoding/json"
	"net/http"

	"bitbucket.org/andyfusniakteam/ecom-api-go/services"
	"github.com/gorilla/mux"
)

type itemRequestBody struct {
	Sku string `json:"sku"`
	Qty int    `json:"qty"`
}

// AddItemToCart adds an item to a given cart
func AddItemToCart(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)

	o := itemRequestBody{}
	err := json.NewDecoder(r.Body).Decode(&o)
	if err != nil {
		http.Error(w, err.Error(), 400)
		return
	}

	item, err := services.AddItemToCart(params["ctid"], o.Sku, o.Qty)
	if err != nil {
		panic(err)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated) // 201 Created
	json.NewEncoder(w).Encode(item)
}
