package app

import (
	"encoding/json"
	"fmt"
	"net/http"

	"bitbucket.org/andyfusniakteam/ecom-api-go/service/firebase"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

type categoryAssoc struct {
	Products []string `json:"products"`
}

type categoryAssocs map[string]categoryAssoc

// UpdateCatalogProductAssocsHandler creates a handler function that overwrites
// a new category association.
func (a *App) UpdateCatalogProductAssocsHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		contextLogger := log.WithContext(ctx)
		contextLogger.Info("App: UpdateCatalogProductAssocsHandler started")

		if r.Body == nil {
			w.WriteHeader(http.StatusBadRequest) // 400 Bad Request
			json.NewEncoder(w).Encode(struct {
				Status  int    `json:"status"`
				Code    string `json:"code"`
				Message string `json:"message"`
			}{
				http.StatusBadRequest,
				ErrCodeBadRequest,
				"Please send a request body",
			})
			return
		}

		// Catalog product associations may only be written if a catalog exists.
		has, err := a.Service.HasCatalog(ctx)
		if err != nil {
			contextLogger.Errorf("a.Service.HasCatalog(ctx) failed %+v", errors.Cause(err))
			w.WriteHeader(http.StatusInternalServerError) // 500 Internal Server Error
			return
		}
		if !has {
			w.WriteHeader(http.StatusConflict) // 409 Conflict
			json.NewEncoder(w).Encode(struct {
				Status  int    `json:"status"`
				Code    string `json:"code"`
				Message string `json:"message"`
			}{
				http.StatusConflict,
				ErrCodeNoCatalog,
				"category product associations can only be updated if a catalog exists.",
			})
			return
		}

		cas := map[string]categoryAssoc{}
		err = json.NewDecoder(r.Body).Decode(&cas)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest) // 400 Bad Request
			json.NewEncoder(w).Encode(struct {
				Status  int    `json:"status"`
				Code    string `json:"code"`
				Message string `json:"message"`
			}{
				http.StatusBadRequest,
				ErrCodeBadRequest,
				err.Error(),
			})
			return
		}
		defer r.Body.Close()

		tree, _ := a.Service.GetCatalog(ctx)
		skus, missingPaths, nonLeafs := validateCatalogAssocs(cas, tree)
		_, missingSKUs, err := a.Service.ProductsExist(ctx, skus)
		if err != nil {
			contextLogger.Errorf("a.Service.ProductsExist(ctx, ...) failed: %+v", errors.Cause(err))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		if missingPaths != nil || nonLeafs != nil || missingSKUs != nil {
			w.WriteHeader(http.StatusConflict) // 409 Conflict
			json.NewEncoder(w).Encode(struct {
				Status  int    `json:"status"`
				Code    string `json:"code"`
				Message string `json:"message"`
				Data    struct {
					MissingPaths []string `json:"missing_paths"`
					NonLeafPaths []string `json:"non_leaf_paths"`
					MissingSKUs  []string `json:"missing_skus"`
				} `json:"data"`
			}{
				http.StatusConflict,
				ErrMissingPathsLeafsSKUs,
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
		for path, a := range cas {
			skus := make([]string, 0, 32)
			for _, s := range a.Products {
				skus = append(skus, s)
			}
			cpas[path] = skus
		}
		err = a.Service.CreateCategoryProductAssocs(ctx, cpas)
		if err != nil {
			contextLogger.Errorf("CreateCategoryProductAssocs(ctx, ...) failed: %+v", err)
			w.WriteHeader(http.StatusInternalServerError) // 500 Internal Server Error
			return
		}
		w.WriteHeader(http.StatusNoContent) // 204 No Content
	}
}

func validateCatalogAssocs(cmap map[string]categoryAssoc, tree *firebase.Category) (skus, missingPaths, nonLeafs []string) {
	skumap := make(map[string]bool)
	for path, ca := range cmap {
		for _, s := range ca.Products {
			if _, ok := skumap[s]; !ok {
				skumap[s] = true
			}
		}
		n := tree.FindNodeByPath(path)
		if n == nil {
			missingPaths = append(missingPaths, path)
		} else if !n.IsLeaf() {
			nonLeafs = append(nonLeafs, path)
		}
	}
	// convert map keys to slice
	skus = make([]string, 0, 128)
	for k := range skumap {
		skus = append(skus, k)
	}
	return skus, missingPaths, nonLeafs
}
