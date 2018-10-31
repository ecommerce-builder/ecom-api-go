package controllers

import (
	"fmt"
	"net/http"

	"bitbucket.org/andyfusniakteam/ecom-api-go/services"
	"github.com/gorilla/mux"
)

// DeleteCartItem adds an item to a given cart
func DeleteCartItem(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)

	count, _ := services.DeleteCartItem(params["ctid"], params["sku"])
	fmt.Println(count)

	if count == 0 {
		w.WriteHeader(http.StatusNotFound) // 404 Not Found
		return
	}

	w.WriteHeader(http.StatusNoContent) // 204 No Content
}
