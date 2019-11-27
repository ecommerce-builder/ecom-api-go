package app

import (
	"encoding/json"
	"net/http"

	service "bitbucket.org/andyfusniakteam/ecom-api-go/service/firebase"
	log "github.com/sirupsen/logrus"
)

// ListPPAssocGroupsHandler creates a handler function that returns a
// list of product to product association groups.
func (a *App) ListPPAssocGroupsHandler() http.HandlerFunc {
	type listPromoRulesResponse struct {
		Object string                  `json:"object"`
		Data   []*service.PPAssocGroup `json:"data"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		contextLogger := log.WithContext(ctx)
		contextLogger.Info("app: ListPPAssocGroupsHandler started")

		ppAssocGroups, err := a.Service.GetPPAssocGroups(ctx)
		if err != nil {
			contextLogger.Errorf("app: a.Service.GetPPAssocGroups(ctx) error: %+v", err)
			w.WriteHeader(http.StatusInternalServerError) // 500 Internal Server Error
			return
		}

		list := listPromoRulesResponse{
			Object: "list",
			Data:   ppAssocGroups,
		}
		w.WriteHeader(http.StatusOK) // 200 OK
		json.NewEncoder(w).Encode(&list)
	}
}
