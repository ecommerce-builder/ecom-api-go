package app

import (
	"encoding/json"
	"net/http"

	service "bitbucket.org/andyfusniakteam/ecom-api-go/service/firebase"
	log "github.com/sirupsen/logrus"
)

type applyCartCouponRequest struct {
	CartID   *string `json:"cart_id"`
	CouponID *string `json:"coupon_id"`
}

func validateApplyCartCouponRequestMemoize() func(*applyCartCouponRequest) (bool, string) {
	return func(request *applyCartCouponRequest) (bool, string) {
		// cart_id attribute
		cartID := request.CartID
		if cartID == nil {
			return false, "cart_id attribute must be set"
		}
		if !IsValidUUID(*cartID) {
			return false, "cart_id attribute must be a valid v4 UUID"
		}

		// coupon_id attribute
		couponID := request.CouponID
		if request.CouponID == nil {
			return false, "coupon_id attribute must be set"
		}
		if !IsValidUUID(*couponID) {
			return false, "coupon_id attribute must be a valid v4 UUID"
		}

		return true, ""
	}
}

// ApplyCartCouponHandler applies a coupon to a cart.
func (a *App) ApplyCartCouponHandler() http.HandlerFunc {
	validate := validateApplyCartCouponRequestMemoize()

	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		contextLogger := log.WithContext(ctx)
		contextLogger.Info("app: CreateCouponHandler called")

		request := applyCartCouponRequest{}
		dec := json.NewDecoder(r.Body)
		dec.DisallowUnknownFields()
		if err := dec.Decode(&request); err != nil {
			clientError(w, http.StatusBadRequest, ErrCodeBadRequest,
				err.Error()) // 400
			return
		}

		valid, message := validate(&request)
		if !valid {
			clientError(w, http.StatusBadRequest, ErrCodeBadRequest, message) // 400
			return
		}

		coupon, err := a.Service.ApplyCouponToCart(ctx, *request.CartID, *request.CouponID)
		if err == service.ErrCartNotFound {
			clientError(w, http.StatusNotFound, ErrCodeCartNotFound,
				"cart not found") // 404
		}
		if err == service.ErrCouponNotFound {
			clientError(w, http.StatusNotFound, ErrCodeCouponNotFound,
				"coupon not found") // 404
			return
		}
		if err == service.ErrCartCouponExists {
			clientError(w, http.StatusConflict, ErrCodeCartCouponExists,
				"coupon already in cart") // 409
			return
		}
		if err == service.ErrCouponNotAtStartDate {
			clientError(w, http.StatusConflict, ErrCodeCouponNotAtStartDate,
				"coupon promotion date has not started") // 409
		}
		if err == service.ErrCouponExpired {
			clientError(w, http.StatusConflict, ErrCodeCouponExpired,
				"coupon expired") // 409
			return
		}
		if err == service.ErrCouponVoid {
			clientError(w, http.StatusConflict, ErrCodeCouponVoid,
				"coupon void") // 409
			return
		}
		if err == service.ErrCouponUsed {
			clientError(w, http.StatusConflict, ErrCodeCouponUsed,
				"coupon has already been used") // 409
			return
		}
		if err != nil {
			contextLogger.Errorf("app: a.Service.ApplyCouponToCart(ctx, crtID=%q, couponID=%q) failed: %+v", *request.CartID, *request.CouponID, err)
			w.WriteHeader(http.StatusInternalServerError) // 500
			return
		}
		w.WriteHeader(http.StatusCreated) // 201
		json.NewEncoder(w).Encode(&coupon)
	}
}
