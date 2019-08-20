package app

import (
	"net/http"

	service "bitbucket.org/andyfusniakteam/ecom-api-go/service/firebase"
	"github.com/go-chi/chi"
	log "github.com/sirupsen/logrus"
)

// DeletePromoRuleHandler creates a handler function that deletes
// a price list by id.
func (a *App) DeletePromoRuleHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		contextLogger := log.WithContext(ctx)
		contextLogger.Info("App: DeletePromoRuleHandler started")

		promoRuleID := chi.URLParam(r, "id")
		if err := a.Service.DeletePromoRule(ctx, promoRuleID); err != nil {
			if err == service.ErrPromoRuleNotFound {
				w.WriteHeader(http.StatusNotFound) // 404 Not Found
				return
			}
			contextLogger.Errorf("app: a.Service.DeletePromoRule(ctx, promoRuleID=%q) error: %+v", promoRuleID, err)
			w.WriteHeader(http.StatusInternalServerError) // 500 Internal Server Error
			return
		}
		w.Header().Del("Content-Type")
		w.WriteHeader(http.StatusNoContent) // 204 No Content
	}
}
