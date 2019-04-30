package app

import (
	"encoding/json"
	"fmt"
	"net/http"

	"bitbucket.org/andyfusniakteam/ecom-api-go/utils/nestedset"
)

// CreateCatalogAssocsHandler 
func (a *App) CreateCatalogAssocsHandler() http.HandlerFunc {

	type catalogProduct struct {
		SKU string `json:"sku"`
	}

	type catalogAssoc struct {
		Path string                `json:"path"`
		Products []catalogProduct `json:"products"`
	}
	
	type catalogAssocs []catalogAssoc

	return func(w http.ResponseWriter, r *http.Request) {
		if r.Body == nil {
			w.WriteHeader(http.StatusBadRequest) // 400 Bad Request
			json.NewEncoder(w).Encode(struct {
				Code    int    `json:"code"`
				Message string `json:"message"`
			}{
				400,
				"Please send a request body",
			})
			return
		}

		cpas := catalogAssocs{}
		err := json.NewDecoder(r.Body).Decode(&cpas)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest) // 400 Bad Request
			json.NewEncoder(w).Encode(struct {
				Code    int    `json:"code"`
				Message string `json:"message"`
			}{
				400,
				err.Error(),
			})
			return
		}
		defer r.Body.Close()

		//for _, ca := range cpas {
		//	fmt.Printf("%+v\n", ca)
		//}

		ns, _ := a.Service.GetCatalog(r.Context()) // ([]*nestedset.NestedSetNode, error)
		
		tree := nestedset.BuildTree(ns)

		fmt.Printf("%#v\n", tree)

		w.WriteHeader(http.StatusCreated) // 201 Created
		//json.NewEncoder(w).Encode(*cpas)
	}
}
