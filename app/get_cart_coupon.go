package app

import (
	"encoding/json"
	"net/http"

	service "bitbucket.org/andyfusniakteam/ecom-api-go/service/firebase"
	"github.com/go-chi/chi"
	log "github.com/sirupsen/logrus"
)

// GetCartCouponHandler returns a http.HandlerFunc that returns a single
// cart coupon.
func (a *App) GetCartCouponHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		contextLogger := log.WithContext(ctx)
		contextLogger.Info("app: GetCartCouponHandler called")

		cartCouponID := chi.URLParam(r, "id")
		cartCoupon, err := a.Service.GetCartCoupon(ctx, cartCouponID)
		if err != nil {
			if err == service.ErrCartCouponNotFound {
				clientError(w, http.StatusNotFound, ErrCodeCartCouponNotFound, "cart coupon not found")
				return
			}
			contextLogger.Errorf("app: a.Service.GetCartCoupon(ctx, cartCouponID=%q) failed: %+v", cartCouponID, err)
			return
		}
		w.WriteHeader(http.StatusOK) // 200 OK
		json.NewEncoder(w).Encode(&cartCoupon)
	}
}
