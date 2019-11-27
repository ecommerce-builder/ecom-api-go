package app

import (
	"net/http"

	service "bitbucket.org/andyfusniakteam/ecom-api-go/service/firebase"
	"github.com/go-chi/chi"
	log "github.com/sirupsen/logrus"
)

// UnapplyCartCouponHandler creates a handler to unapply a cart coupon.
func (a *App) UnapplyCartCouponHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		contextLogger := log.WithContext(ctx)
		contextLogger.Info("app: UnapplyCartCouponHandler started")

		cartCouponID := chi.URLParam(r, "id")
		if cartCouponID == "" {
			clientError(w, http.StatusBadRequest, ErrCodeBadRequest,
				"path parameter id must be set") // 400
			return
		}
		if !IsValidUUID(cartCouponID) {
			clientError(w, http.StatusBadRequest, ErrCodeBadRequest,
				"path parameter id must be a valid v4 uuid") // 400
			return
		}

		err := a.Service.UnapplyCartCoupon(ctx, cartCouponID)
		if err == service.ErrCartCouponNotFound {
			clientError(w, http.StatusNotFound, ErrCodeCartCouponNotFound,
				"cart coupon not found") // 404
			return
		}
		if err != nil {
			contextLogger.Errorf("app: a.Service.UnapplyCartCoupon(ctx, cartCouponID=%q) error: %+v", cartCouponID, err)
			w.WriteHeader(http.StatusInternalServerError) // 500
			return
		}

		w.Header().Del("Content-Type")
		w.Header().Set("Content-Length", "0")
		w.WriteHeader(http.StatusNoContent) // 204
	}
}
