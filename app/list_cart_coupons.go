package app

import (
	"encoding/json"
	"net/http"

	service "bitbucket.org/andyfusniakteam/ecom-api-go/service/firebase"
	log "github.com/sirupsen/logrus"
)

// ListCartCouponsHandler returns a http.HandlerFunc that returns
// a list of cart coupons relations.
func (a *App) ListCartCouponsHandler() http.HandlerFunc {
	type response struct {
		Object string                `json:"object"`
		Data   []*service.CartCoupon `json:"data"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		contextLogger := log.WithContext(ctx)
		contextLogger.Info("app: ListCartCouponsHandler started")

		cartID := r.URL.Query().Get("cart_id")
		if cartID == "" {
			clientError(w, http.StatusBadRequest, ErrCodeBadRequest,
				"cart_id query param must be set") // 400
			return
		}
		if !IsValidUUID(cartID) {
			clientError(w, http.StatusBadRequest, ErrCodeBadRequest,
				"cart_id query param must be a valid v4 UUID") // 400
			return
		}

		cartCoupons, err := a.Service.GetCartCoupons(ctx, cartID)
		if err != nil {
			contextLogger.Errorf("app: a.Service.GetCartCoupons(ctx, cartID) failed: %+v", err)
			w.WriteHeader(http.StatusInternalServerError) // 500
			return
		}

		list := response{
			Object: "list",
			Data:   cartCoupons,
		}
		w.WriteHeader(http.StatusOK) // 200
		json.NewEncoder(w).Encode(&list)
	}
}
