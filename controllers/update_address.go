package controllers

import (
	"encoding/json"
	"net/http"

	"bitbucket.org/andyfusniakteam/ecom-api-go/models"
)

// UpdateAddress handler
func UpdateAddress(w http.ResponseWriter, r *http.Request) {
	//params := mux.Vars(r)

	//services.UpdateAddress()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusNoContent) // 204 No Content
	json.NewEncoder(w).Encode(models.Customer{})
}
