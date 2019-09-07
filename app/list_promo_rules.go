package app

import (
	"encoding/json"
	"net/http"

	service "bitbucket.org/andyfusniakteam/ecom-api-go/service/firebase"
	log "github.com/sirupsen/logrus"
)

// ListPromoRulesHandler creates a handler function that returns a
// list of promotion rules.
func (a *App) ListPromoRulesHandler() http.HandlerFunc {
	type listPromoRulesResponse struct {
		Object string               `json:"object"`
		Data   []*service.PromoRule `json:"data"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		contextLogger := log.WithContext(ctx)
		contextLogger.Info("app: ListPromoRulesHandler started")

		promoRules, err := a.Service.GetPromoRules(ctx)
		if err != nil {
			contextLogger.Errorf("app: GetPromoRules(ctx) error: %+v", err)
			w.WriteHeader(http.StatusInternalServerError) // 500 Internal Server Error
			return
		}

		list := listPromoRulesResponse{
			Object: "list",
			Data:   promoRules,
		}
		w.WriteHeader(http.StatusOK) // 200 OK
		json.NewEncoder(w).Encode(&list)
	}
}
