package app

import (
	"encoding/json"
	"fmt"
	"net/http"

	"bitbucket.org/andyfusniakteam/ecom-api-go/utils/nestedset"
)

type catalogProduct struct {
	SKU string `json:"sku"`
}

type catalogAssoc struct {
	Path     string           `json:"path"`
	Products []catalogProduct `json:"products"`
}

type catalogAssocs []catalogAssoc

// CreateCatalogAssocsHandler creates a handler function that creates a new catalog association.
func (a *App) CreateCatalogAssocsHandler() http.HandlerFunc {
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

		cas := catalogAssocs{}
		err := json.NewDecoder(r.Body).Decode(&cas)
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

		tree, _ := a.Service.GetCatalog(r.Context()) // ([]*nestedset.NestedSetNode, error)
		fmt.Printf("%#v\n", tree)

		skus, missingPaths, nonLeafs := validateCatalogAssocs(cas, tree)
		if skus != nil || missingPaths != nil || nonLeafs != nil {
			fmt.Printf("SKUs in payload: %v\n", skus)
			fmt.Printf("Missing paths: %v\n", missingPaths)
			fmt.Printf("Non-leaf paths: %v\n", nonLeafs)
		} else {
			fmt.Println("all is well")
		}

		w.WriteHeader(http.StatusCreated) // 201 Created
		//json.NewEncoder(w).Encode(*cpas)
	}
}

func validateCatalogAssocs(cas catalogAssocs, tree *nestedset.Node) (skus, missingPaths, nonLeafs []string) {
	skumap := make(map[string]bool)
	for _, ca := range cas {
		for _, cp := range ca.Products {
			if _, ok := skumap[cp.SKU]; !ok {
				skumap[cp.SKU] = true
			}
		}
		n := tree.FindNodeByPath(ca.Path)
		if n == nil {
			missingPaths = append(missingPaths, ca.Path)
		} else if n.IsLeaf() {
			nonLeafs = append(nonLeafs, ca.Path)
		}
	}

	// convert map keys to slice
	skus = make([]string, 0, 128)
	for k := range skumap {
		skus = append(skus, k)
	}
	return skus, missingPaths, nonLeafs
}
