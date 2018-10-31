package controllers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"bitbucket.org/andyfusniakteam/ecom-api-go/services"
	"github.com/gorilla/mux"
)

type addressRequestBody struct {
	Typ         string `json:"typ"`
	ContactName string `json:"contact_name"`
	Addr1       string `json:"address_line_1`
	Addr2       string `json:"address_line_2`
	City        string `json:"city"`
	County      string `json:"county"`
	Postcode    string `json:"postcode"`
	Country     string `json:"country"`
}

// CreateAddress handler
func CreateAddress(w http.ResponseWriter, r *http.Request) {
	if r.Body == nil {
		http.Error(w, "Please send a request body", 400)
		return
	}

	params := mux.Vars(r)

	fmt.Println(params["cid"])

	o := addressRequestBody{}
	err := json.NewDecoder(r.Body).Decode(&o)
	if err != nil {
		http.Error(w, err.Error(), 400)
		return
	}

	fmt.Println(o)

	address := services.CreateAddress(params["cid"], o.Typ, o.ContactName, o.Addr1, o.Addr2, o.City, o.County, o.Postcode, o.Country)

	w.WriteHeader(201)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(address)
}
