package controllers

import (
	"encoding/json"
	"net/http"

	models "bitbucket.org/andyfusniakteam/ecom-api-go/models"
)

func DeleteAddress(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(models.Customer{})
}
