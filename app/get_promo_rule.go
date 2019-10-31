package app

import (
	"encoding/json"
	"net/http"

	service "bitbucket.org/andyfusniakteam/ecom-api-go/service/firebase"
	"github.com/go-chi/chi"
	log "github.com/sirupsen/logrus"
)

// GetPromoRuleHandler creates a handler function that returns a
// price list.
func (a *App) GetPromoRuleHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		contextLogger := log.WithContext(ctx)
		contextLogger.Info("app: GetPromoRuleHandler called")

		promoRuleID := chi.URLParam(r, "id")
		if !IsValidUUID(promoRuleID) {
			contextLogger.Infof("app: url path param is invalid (%q) - 400 Bad Request",
				promoRuleID)
			clientError(w, http.StatusBadRequest, ErrCodeBadRequest,
				"url path parameter id must be a valid v4 uuid") // 400
			return
		}
		promoRule, err := a.Service.GetPromoRule(ctx, promoRuleID)
		if err == service.ErrPromoRuleNotFound {
			contextLogger.Infof("app: promo rule %q not found - 404 Not found", promoRuleID)
			clientError(w, http.StatusNotFound, ErrCodePromoRuleNotFound,
				"promo rule not found") // 404
			return
		}
		if err != nil {
			contextLogger.Errorf("app: a.Service.GetPromoRule(ctx, promoRuleID=%q): %+v",
				promoRuleID, err)
			w.WriteHeader(http.StatusInternalServerError) // 500
			return
		}

		contextLogger.WithFields(log.Fields{
			"promoRule": promoRule,
		}).Infof("app: 200 OK")
		w.WriteHeader(http.StatusOK) // 200
		json.NewEncoder(w).Encode(&promoRule)
	}
}
