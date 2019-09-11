package app

import (
	"encoding/json"
	"net/http"
	"regexp"

	service "bitbucket.org/andyfusniakteam/ecom-api-go/service/firebase"
	log "github.com/sirupsen/logrus"
)

type createCouponRequestBody struct {
	CouponCode  *string `json:"coupon_code"`
	PromoRuleID *string `json:"promo_rule_id"`
	Reusable    *bool   `json:"reusable"`
}

func validateCreateCouponRequestMemoize() func(*createCouponRequestBody) (bool, string) {
	couponCodeRegExp, err := regexp.Compile("^[A-Za-z0-9]{1,32}$")
	if err != nil {
		log.Errorf("failed to compile regexp for CreateCouponHandler: %v", err)
	}

	return func(request *createCouponRequestBody) (bool, string) {
		// coupon_code attribute
		if request.CouponCode == nil {
			return false, "coupon_code attribute must be set"
		}
		if !couponCodeRegExp.Match([]byte(*request.CouponCode)) {
			return false, "coupon_code attribute must be A-Za-z0-9 only"
		}

		// promo_rule_id attribute
		if request.PromoRuleID == nil {
			return false, "promo_rule_id attribute must be set"
		}
		if !IsValidUUID(*request.PromoRuleID) {
			return false, "promo_rule_id attribute must be a valid v4 UUID"
		}

		// reusable attribute
		if request.Reusable == nil {
			return false, "reusable attribute must be set"
		}

		return true, ""
	}
}

// CreateCouponHandler creates a new coupon.
func (a *App) CreateCouponHandler() http.HandlerFunc {
	validateCreateCouponRequest := validateCreateCouponRequestMemoize()

	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		contextLogger := log.WithContext(ctx)
		contextLogger.Info("app: CreateCouponHandler called")

		request := createCouponRequestBody{}
		dec := json.NewDecoder(r.Body)
		dec.DisallowUnknownFields()
		if err := dec.Decode(&request); err != nil {
			clientError(w, http.StatusBadRequest, ErrCodeBadRequest, err.Error())
			return
		}

		valid, message := validateCreateCouponRequest(&request)
		if !valid {
			clientError(w, http.StatusBadRequest, ErrCodeBadRequest, message)
			return
		}

		coupon, err := a.Service.CreateCoupon(ctx, *request.CouponCode, *request.PromoRuleID, *request.Reusable)
		if err != nil {
			if err == service.ErrCouponExists {
				clientError(w, http.StatusConflict, ErrCodeCouponExists, "coupon with this coupon code already exists")
				return
			} else if err == service.ErrPromoRuleNotFound {
				clientError(w, http.StatusNotFound, ErrCodePromoRuleNotFound, "promo rule not found")
			}

			contextLogger.Errorf("a.Service.CreateCoupon(ctx, couponCode=%q, promoRuleID=%q, reusable=%t) failed: %+v", *request.CouponCode, *request.PromoRuleID, *request.Reusable, err)
			w.WriteHeader(http.StatusInternalServerError) // 500 Internal Server Error
		}
		w.WriteHeader(http.StatusCreated) // 201 Created
		json.NewEncoder(w).Encode(&coupon)
	}
}
