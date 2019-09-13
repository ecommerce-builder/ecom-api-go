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
			clientError(w, http.StatusBadRequest, ErrCodeBadRequest, "URL param id must be set")
			return
		}
		if !IsValidUUID(cartCouponID) {
			clientError(w, http.StatusBadRequest, ErrCodeBadRequest, "URL param id must be a valid v4 UUID")
			return
		}

		if err := a.Service.UnapplyCartCoupon(ctx, cartCouponID); err != nil {
			if err == service.ErrCartCouponNotFound {
				clientError(w, http.StatusNotFound, ErrCodeCartCouponNotFound, "cart coupon not found")
				return
			}
			contextLogger.Errorf("app: a.Service.UnapplyCartCoupon(ctx, cartCouponID) error: %+v", cartCouponID, err)
			w.WriteHeader(http.StatusInternalServerError) // 500 Internal Server Error
			return
		}

		w.Header().Del("Content-Type")
		w.WriteHeader(http.StatusNoContent) // 204 No Content
	}
}
