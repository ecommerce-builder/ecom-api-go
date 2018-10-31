package controllers

import (
	"net/http"

	"bitbucket.org/andyfusniakteam/ecom-api-go/services"
	"github.com/gorilla/mux"
)

// EmptyCartItems empty the cart of all items. This does not remove coupons
func EmptyCartItems(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)

	services.EmptyCartItems(params["ctid"])

	w.WriteHeader(http.StatusNoContent) // 204 No Content
}
