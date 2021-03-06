package app

import (
	"net/http"

	service "bitbucket.org/andyfusniakteam/ecom-api-go/service/firebase"
	"github.com/go-chi/chi"
	log "github.com/sirupsen/logrus"
)

// DeleteCouponHandler creates a handler function that deletes
// a price list by id.
func (a *App) DeleteCouponHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		contextLogger := log.WithContext(ctx)
		contextLogger.Info("app: DeleteCouponHandler started")

		couponID := chi.URLParam(r, "id")
		if !IsValidUUID(couponID) {
			clientError(w, http.StatusBadRequest, ErrCodeBadRequest,
				"path parameter id must be a valid v4 uuid") // 400
			return
		}
		err := a.Service.DeleteCoupon(ctx, couponID)
		if err == service.ErrCouponNotFound {
			clientError(w, http.StatusNotFound, ErrCodeCouponNotFound,
				"coupon not found") // 404
			return
		}
		if err == service.ErrCouponInUse {
			clientError(w, http.StatusConflict, ErrCodeCouponInUse,
				"coupon is already in use, consider making it void instead") // 409
			return
		}
		if err != nil {
			contextLogger.Errorf("app: a.Service.DeleteCoupon(ctx, couponID=%q) error: %+v", couponID, err)
			w.WriteHeader(http.StatusInternalServerError) // 500
			return
		}
		w.Header().Del("Content-Type")
		w.WriteHeader(http.StatusNoContent) // 204
	}
}
