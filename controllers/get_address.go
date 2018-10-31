package controllers

import (
	"encoding/json"
	"net/http"

	"bitbucket.org/andyfusniakteam/ecom-api-go/services"
	"github.com/gorilla/mux"
)

// GetAddress handler
func GetAddress(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)

	a := services.GetAddress(params["aid"])

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK) // 200 OK
	json.NewEncoder(w).Encode(a)
}
