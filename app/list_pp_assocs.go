package app

import (
	"encoding/json"
	"net/http"

	service "bitbucket.org/andyfusniakteam/ecom-api-go/service/firebase"
	log "github.com/sirupsen/logrus"
)

// ListPPAssocsHandler creates a handler function that returns product
// to product associations.
func (a *App) ListPPAssocsHandler() http.HandlerFunc {
	type response struct {
		Object string             `json:"object"`
		Data   []*service.PPAssoc `json:"data"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		contextLogger := log.WithContext(ctx)
		contextLogger.Info("app: ListPPAssocsHandler called")

		ppAssocGroupID := r.URL.Query().Get("pp_assoc_group_id")
		productFromID := r.URL.Query().Get("product_from_id")
		if productFromID != "" {
			if !IsValidUUID(productFromID) {
				clientError(w, http.StatusBadRequest, ErrCodeBadRequest, "optional query parameter product_from_id must be a valid v4 UUID")
				return
			}
		}
		if !IsValidUUID(ppAssocGroupID) {
			clientError(w, http.StatusBadRequest, ErrCodeBadRequest, "pp_assoc_group_id must be a valid v4 UUID")
			return
		}

		ppAssocs, err := a.Service.GetPPAssocs(ctx, ppAssocGroupID, productFromID)
		if err == service.ErrPPAssocGroupNotFound {
			clientError(w, http.StatusNotFound, ErrCodePPAssocGroupNotFound, "product to product association group not found")
			return
		}
		if err == service.ErrProductNotFound {
			clientError(w, http.StatusNotFound, ErrCodeProductNotFound, "product_from_id product not found")
			return
		}
		if err != nil {
			contextLogger.Errorf("app: a.Service.GetPPAssocs(ctx) failed: %+v", err)
			w.WriteHeader(http.StatusInternalServerError) // 500 Internal Server Error
			return
		}

		list := response{
			Object: "list",
			Data:   ppAssocs,
		}
		w.WriteHeader(http.StatusOK) // 200 OK
		json.NewEncoder(w).Encode(&list)
	}
}
