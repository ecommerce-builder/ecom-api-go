package app

import (
	"encoding/json"
	"fmt"
	"net/http"

	service "bitbucket.org/andyfusniakteam/ecom-api-go/service/firebase"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

// GetProductCategoryAssocsHandler creates a handler to return all
// product to category associations.
func (a *App) GetProductCategoryAssocsHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		contextLogger := log.WithContext(ctx)
		contextLogger.Info("App: GetCategoryAssocsHandler started")

		key := r.URL.Query().Get("key")
		if key == "" {
			key = "id"
		}
		if key != "id" && key != "path" {
			w.WriteHeader(http.StatusBadRequest) // 400 Bad Request
			json.NewEncoder(w).Encode(struct {
				Status  int    `json:"status"`
				Code    string `json:"code"`
				Message string `json:"message"`
			}{
				http.StatusBadRequest,
				ErrCodeBadRequest,
				"if set, the query parameter key must contain a value of id or path",
			})
			return
		}

		cpo, err := a.Service.GetProductCategoryRelations(ctx, key)
		if err != nil {
			contextLogger.Errorf("service GetProductCategoryAssocs(ctx) error: %v", err)
			w.WriteHeader(http.StatusInternalServerError) // 500 Internal Server Error
			json.NewEncoder(w).Encode(struct {
				Status  int    `json:"status"`
				Code    string `json:"code"`
				Message string `json:"message"`
			}{
				http.StatusInternalServerError,
				ErrCodeInternalServerError,
				err.Error(),
			})
			return
		}
		w.WriteHeader(http.StatusOK) // 200 OK
		json.NewEncoder(w).Encode(cpo)
	}
}

// PurgeProductCategoryRelationsHandler returns an handler that deletes all category
// to product associations.
func (a *App) PurgeProductCategoryRelationsHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		contextLogger := log.WithContext(ctx)
		contextLogger.Info("app: DeleteProductCategoryRelationsHandler started")

		err := a.Service.DeleteAllProductCategoryRelations(ctx)
		if err != nil {
			contextLogger.Errorf("app: DeleteAllCategoryAssocs(ctx) error: %+v", err)
			w.WriteHeader(http.StatusInternalServerError) // 500 Internal Server Error
			return
		}
		w.Header().Del("Content-Type")
		w.WriteHeader(http.StatusNoContent) // 204 No Content
	}
}

type productCategory struct {
	ProductID string `json:"product_id"`
}

type categoryAssoc struct {
	Products []*productCategory `json:"products"`
}

type categoryAssocs map[string]*categoryAssoc

func validateCatalogAssocs(cmap map[string]*categoryAssoc, tree *service.CategoryNode) (pids, missingPaths, nonLeafs []string) {
	productIDMap := make(map[string]bool)
	for path, ca := range cmap {
		for _, s := range ca.Products {
			if _, ok := productIDMap[s.ProductID]; !ok {
				productIDMap[s.ProductID] = true
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
	pids = make([]string, 0, 128)
	for k := range productIDMap {
		pids = append(pids, k)
	}
	return pids, missingPaths, nonLeafs
}

// UpdateProductCategoryAssocsHandler creates a handler function that overwrites
// a new category association.
func (a *App) UpdateProductCategoryAssocsHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		contextLogger := log.WithContext(ctx)
		contextLogger.Info("App: UpdateProductCategoryAssocsHandler started")

		xKeyedBy := r.Header.Get("X-Keyed-By")
		if xKeyedBy == "" {
			xKeyedBy = "id"
		}
		if xKeyedBy != "id" && xKeyedBy != "path" {
			w.WriteHeader(http.StatusBadRequest) // 400 Bad Request
			json.NewEncoder(w).Encode(struct {
				Status  int    `json:"status"`
				Code    string `json:"code"`
				Message string `json:"message"`
			}{
				http.StatusBadRequest,
				ErrCodeBadRequest,
				"X-Keyed-By header must to be set to either id or path",
			})
			return
		}

		// TODO: Using the category id as a key for associating a list products
		// is not yet implemented. Use X-Keyed-By: path for now.
		if xKeyedBy == "id" {
			w.WriteHeader(http.StatusNotImplemented) // 501 Not Implemented
			return
		}

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

		// Category product associations may only be written if at least one
		// category exists.
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
				"category product associations can only be updated if a categories exists.",
			})
			return
		}

		categoryAssocs := map[string]*categoryAssoc{}
		err = json.NewDecoder(r.Body).Decode(&categoryAssocs)
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

		tree, err := a.Service.GetCategoriesTree(ctx)
		if err != nil {
			contextLogger.Errorf("a.Service.GetCatalog(ctx) failed: %+v", errors.Cause(err))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		pids, missingPaths, nonLeafs := validateCatalogAssocs(categoryAssocs, tree)
		_, missingProductIDs, err := a.Service.ProductsExist(ctx, pids)
		if err != nil {
			contextLogger.Errorf("a.Service.ProductsExist(ctx, ...) failed: %+v", errors.Cause(err))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		if missingPaths != nil || nonLeafs != nil || missingProductIDs != nil {
			w.WriteHeader(http.StatusConflict) // 409 Conflict
			json.NewEncoder(w).Encode(struct {
				Status  int    `json:"status"`
				Code    string `json:"code"`
				Message string `json:"message"`
				Data    struct {
					MissingPaths      []string `json:"missing_paths"`
					NonLeafPaths      []string `json:"non_leaf_paths"`
					MissingProductIDs []string `json:"missing_product_ids"`
				} `json:"data"`
			}{
				http.StatusConflict,
				ErrMissingPathsLeafsProductIDs,
				fmt.Sprintf("Missing paths: %v Non-leaf paths: %v Missing Product IDs: %v", missingPaths, nonLeafs, missingProductIDs),
				struct {
					MissingPaths      []string `json:"missing_paths"`
					NonLeafPaths      []string `json:"non_leaf_paths"`
					MissingProductIDs []string `json:"missing_product_ids"`
				}{
					MissingPaths:      missingPaths,
					NonLeafPaths:      nonLeafs,
					MissingProductIDs: missingProductIDs,
				},
			})
			return
		}

		cpas := map[string][]string{}
		for path, a := range categoryAssocs {
			pids := make([]string, 0, 32)
			for _, s := range a.Products {
				pids = append(pids, s.ProductID)
			}
			cpas[path] = pids
		}
		err = a.Service.CreateProductCategoryRelations(ctx, cpas)
		if err != nil {
			contextLogger.Errorf("CreateProductCategoryAssocs(ctx, ...) failed: %+v", err)
			w.WriteHeader(http.StatusInternalServerError) // 500 Internal Server Error
			return
		}
		w.WriteHeader(http.StatusNoContent) // 204 No Content
	}
}
