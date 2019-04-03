package app

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/go-chi/chi"
)

// CreateAddressHandler
func (a *App) CreateAddressHandler() http.HandlerFunc {
	type addressRequestBody struct {
		Typ         string  `json:"typ"`
		ContactName string  `json:"contact_name"`
		Addr1       string  `json:"addr1"`
		Addr2       *string `json:"addr2"`
		City        string  `json:"city"`
		County      *string `json:"county"`
		Postcode    string  `json:"postcode"`
		Country     string  `json:"country"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		if r.Body == nil {
			http.Error(w, "Please send a request body", 400)
			return
		}

		cuuid := chi.URLParam(r, "cuuid")

		o := addressRequestBody{}
		err := json.NewDecoder(r.Body).Decode(&o)
		if err != nil {
			http.Error(w, err.Error(), 400)
			return
		}

		address, err := a.Service.CreateAddress(r.Context(), cuuid, o.Typ, o.ContactName, o.Addr1, o.Addr2, o.City, o.County, o.Postcode, o.Country)
		if err != nil {
			fmt.Fprintf(os.Stderr, "service CreateAddress(%s, ...) error: %v", cuuid, err)
			return
		}
		w.WriteHeader(http.StatusCreated) // 201 Created
		json.NewEncoder(w).Encode(*address)
	}
}