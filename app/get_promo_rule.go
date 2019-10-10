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
			clientError(w, http.StatusBadRequest, ErrCodeBadRequest, "URL parameter id must be a valid v4 UUID")
			return
		}
		promoRule, err := a.Service.GetPromoRule(ctx, promoRuleID)
		if err != nil {
			if err == service.ErrPromoRuleNotFound {
				clientError(w, http.StatusNotFound, ErrCodePromoRuleNotFound, "promo rule not found")
				return
			}
			contextLogger.Errorf("app: a.Service.GetPromoRule(ctx, promoRuleID=%q)", promoRuleID)
			w.WriteHeader(http.StatusInternalServerError) // 500 Internal Server Error
		}
		w.WriteHeader(http.StatusOK) // 200 OK
		json.NewEncoder(w).Encode(&promoRule)
	}
}
