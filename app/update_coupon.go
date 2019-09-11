package app

import (
	"encoding/json"
	"net/http"

	service "bitbucket.org/andyfusniakteam/ecom-api-go/service/firebase"
	"github.com/go-chi/chi"
	log "github.com/sirupsen/logrus"
)

type updateCouponRequest struct {
	Void *bool `json:"Void"`
}

func validateUpdateCouponRequest(request *updateCouponRequest) (bool, string) {
	if request.Void == nil {
		return false, "void attribute must be set"
	}

	return true, ""
}

// UpdateCouponHandler returns a http.HandlerFunc that patches a coupon.
func (a *App) UpdateCouponHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		contextLogger := log.WithContext(ctx)
		contextLogger.Info("app: UpdateCouponHandler called")

		// parse request body
		// example
		// { "onhold": 4 }
		var request updateCouponRequest
		dec := json.NewDecoder(r.Body)
		dec.DisallowUnknownFields()
		if err := dec.Decode(&request); err != nil {
			clientError(w, http.StatusBadRequest, ErrCodeBadRequest, err.Error())
			return
		}

		ok, message := validateUpdateCouponRequest(&request)
		if !ok {
			clientError(w, http.StatusBadRequest, ErrCodeBadRequest, message)
			return
		}

		// validate URL parameter
		couponID := chi.URLParam(r, "id")
		if !IsValidUUID(couponID) {
			clientError(w, http.StatusBadRequest, ErrCodeBadRequest, "URL parameter must be a valid v4 UUID")
			return
		}

		coupon, err := a.Service.UpdateCoupon(ctx, couponID, request.Void)
		if err != nil {
			if err == service.ErrCouponNotFound {
				clientError(w, http.StatusNotFound, ErrCodeCouponNotFound, "coupon not found")
				return
			}
			contextLogger.Errorf("app: a.Service.UpdateCoupon(ctx, couponID=%q, void=%v) failed: %+v", request.Void, err)
			w.WriteHeader(http.StatusInternalServerError) // 500 Internal Server Error
			return
		}

		w.WriteHeader(http.StatusOK) // 200 OK
		json.NewEncoder(w).Encode(&coupon)
	}
}
