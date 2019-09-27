package app

import (
	"encoding/json"
	"net/http"

	service "bitbucket.org/andyfusniakteam/ecom-api-go/service/firebase"
	log "github.com/sirupsen/logrus"
)

// ProductToSet is a structured list of products
type ProductToSet struct {
	Object string                     `json:"object"`
	Data   []*service.ProductToUpdate `json:"data"`
}

type batchUpdateRequest struct {
	PPAssocsGroupID *string       `json:"pp_assoc_group_id"`
	ProductFromID   *string       `json:"product_from_id"`
	ProductToSet    *ProductToSet `json:"product_to_set"`
}

func validateBatchUpdatePPAssocsRequestMemoize() func(*batchUpdateRequest) (bool, string) {
	return func(request *batchUpdateRequest) (bool, string) {
		// pp_assocs_group_id attribute
		if request.PPAssocsGroupID == nil {
			return false, "pp_assocs_group_id attribute must be set"
		}
		if !IsValidUUID(*request.PPAssocsGroupID) {
			return false, "pp_assocs_group_id must be a valid v4 UUID"
		}

		// product_from attribute_id
		if request.ProductFromID == nil {
			return false, "product_from_id attribute must be set"
		}
		if !IsValidUUID(*request.ProductFromID) {
			return false, "product_from_id must be a valid v4 UUID"
		}

		// product_to_set attribute
		if request.ProductToSet == nil {
			return false, "product_to_set attribute must be set to a list of products"
		}

		return true, ""
	}
}

// BatchUpdatePPAssocsHandler returns a http.HandlerFunc that batch
// updates product to product associations.
func (a *App) BatchUpdatePPAssocsHandler() http.HandlerFunc {
	validate := validateBatchUpdatePPAssocsRequestMemoize()

	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		contextLogger := log.WithContext(ctx)
		contextLogger.Info("app: BatchUpdatePPAssocsHandler called")

		var request batchUpdateRequest
		dec := json.NewDecoder(r.Body)
		dec.DisallowUnknownFields()
		if err := dec.Decode(&request); err != nil {
			clientError(w, http.StatusBadRequest, ErrCodeBadRequest, err.Error())
			return
		}

		ok, message := validate(&request)
		if !ok {
			clientError(w, http.StatusBadRequest, ErrCodeBadRequest, message)
			return
		}

		err := a.Service.BatchUpdatePPAssocs(ctx, *request.PPAssocsGroupID, *request.ProductFromID, request.ProductToSet.Data)
		if err != nil {
			contextLogger.Errorf("app: a.Service.BatchUpdatePPAssocs(ctx, ppAssocsGroupID=%q, productFrom=%q, productToSet=%v) failed: %+v", *request.PPAssocsGroupID, *request.ProductFromID, *request.ProductToSet, err)
			w.WriteHeader(http.StatusInternalServerError) // 500 Internal Server Error
			return
		}

		w.Header().Del("Content-Type")
		w.Header().Set("Content-Length", "0")
		w.WriteHeader(http.StatusNoContent) // 204 No Content
	}
}
