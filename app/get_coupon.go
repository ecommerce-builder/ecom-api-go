package app

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi"

	service "bitbucket.org/andyfusniakteam/ecom-api-go/service/firebase"
	log "github.com/sirupsen/logrus"
)

// GetCouponHandler creates a handler function that returns a
// price list.
func (a *App) GetCouponHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		contextLogger := log.WithContext(ctx)
		contextLogger.Info("app: GetCouponHandler called")

		couponCodeID := chi.URLParam(r, "id")
		if !IsValidUUID(couponCodeID) {
			clientError(w, http.StatusBadRequest, ErrCodeBadRequest, "url parameter must be a valid v4 UUID")
			return
		}

		coupon, err := a.Service.GetCoupon(ctx, couponCodeID)
		if err != nil {
			if err == service.ErrCouponNotFound {
				clientError(w, http.StatusNotFound, ErrCodeCouponNotFound, "coupon not found")
				return
			}
			contextLogger.Errorf("app: a.Service.GetCoupon(ctx, couponCodeID=%q) failed: %+v", couponCodeID, err)
			w.WriteHeader(http.StatusInternalServerError) // 500 Internal Server Error
			return
		}
		w.WriteHeader(http.StatusOK) // 200 OK
		json.NewEncoder(w).Encode(&coupon)
	}
}
