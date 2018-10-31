package controllers

import (
	"net/http"

	"bitbucket.org/andyfusniakteam/ecom-api-go/services"
	"github.com/gorilla/mux"
)

// DeleteAddress handler
func DeleteAddress(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)

	err := services.DeleteAddress(params["aid"])
	if err != nil {
		panic(err)
	}
	w.WriteHeader(204)
}
