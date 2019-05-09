package app

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"bitbucket.org/andyfusniakteam/ecom-api-go/utils/nestedset"
	"github.com/pkg/errors"
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
		ctx := r.Context()
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

		// Catalog product assoications may only be written if none exist.
		has, err := a.Service.HasCatalogProductAssocs(ctx)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%+v", errors.Cause(err))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		if has {
			w.WriteHeader(http.StatusConflict) // 409 Conflict
			json.NewEncoder(w).Encode(struct {
				Code    int    `json:"code"`
				Message string `json:"message"`
			}{
				400,
				"Catalog product assoications already exist. Call OpPurgeCatalogAssocs first.",
			})
			return
		}

		cas := catalogAssocs{}
		err = json.NewDecoder(r.Body).Decode(&cas)
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

		tree, _ := a.Service.GetCatalog(ctx)
		skus, missingPaths, nonLeafs := validateCatalogAssocs(cas, tree)
		_, missingSKUs, err := a.Service.ProductsExist(ctx, skus)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%+v", errors.Cause(err))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		if missingPaths != nil || nonLeafs != nil || missingSKUs != nil {
			w.WriteHeader(http.StatusConflict) // 409 Conflict
			json.NewEncoder(w).Encode(struct {
				Code    int    `json:"code"`
				Message string `json:"message"`
				Data    struct {
					MissingPaths []string `json:"missing_paths"`
					NonLeafPaths []string `json:"non_leaf_paths"`
					MissingSKUs  []string `json:"missing_skus"`
				} `json:"data"`
			}{
				409,
				fmt.Sprintf("Missing paths: %v Non-leaf paths: %v Missing SKUs: %v", missingPaths, nonLeafs, missingSKUs),
				struct {
					MissingPaths []string `json:"missing_paths"`
					NonLeafPaths []string `json:"non_leaf_paths"`
					MissingSKUs  []string `json:"missing_skus"`
				}{
					MissingPaths: missingPaths,
					NonLeafPaths: nonLeafs,
					MissingSKUs:  missingSKUs,
				},
			})
			return
		}

		cpas := map[string][]string{}
		for _, a := range cas {
			skus := make([]string, 0, 32)
			for _, cp := range a.Products {
				skus = append(skus, cp.SKU)
			}
			cpas[a.Path] = skus
		}
		err = a.Service.BatchCreateCatalogProductAssocs(ctx, cpas)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%+v", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusNoContent) // 204 No Content
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
		} else if !n.IsLeaf() {
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
