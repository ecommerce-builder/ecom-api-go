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
			contextLogger.Warnf("app: 400 Bad Request - url parameter must be a valid v4 uuid (got %q)",
				couponCodeID)
			clientError(w, http.StatusBadRequest, ErrCodeBadRequest,
				"url parameter must be a valid v4 uuid") // 400
			return
		}

		coupon, err := a.Service.GetCoupon(ctx, couponCodeID)
		if err == service.ErrCouponNotFound {
			contextLogger.Infof("app: 404 Not found - coupon id %q not found",
				couponCodeID)
			clientError(w, http.StatusNotFound, ErrCodeCouponNotFound,
				"coupon not found") // 404
			return
		}
		if err != nil {
			contextLogger.Errorf("app: 500 Internal Server Error a.Service.GetCoupon(ctx, couponCodeID=%q) failed: %+v",
				couponCodeID, err)
			w.WriteHeader(http.StatusInternalServerError) // 500
			return
		}

		w.WriteHeader(http.StatusOK) // 200 OK
		json.NewEncoder(w).Encode(&coupon)
		contextLogger.Infof("app: 200 OK - returning coupon data id=%q", couponCodeID)
	}
}
