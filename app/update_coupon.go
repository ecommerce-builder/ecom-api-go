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
		// { "void": true }
		var request updateCouponRequest
		dec := json.NewDecoder(r.Body)
		dec.DisallowUnknownFields()
		if err := dec.Decode(&request); err != nil {
			contextLogger.Warnf("app: 400 Bad Request: %v",
				err.Error())
			clientError(w, http.StatusBadRequest, ErrCodeBadRequest,
				err.Error()) // 400
			return
		}

		ok, message := validateUpdateCouponRequest(&request)
		if !ok {
			contextLogger.Warnf("app: 400 Bad Request - validation failed - %v",
				message)
			clientError(w, http.StatusBadRequest,
				ErrCodeBadRequest, message) // 400
			return
		}

		// validate URL parameter
		couponID := chi.URLParam(r, "id")
		if !IsValidUUID(couponID) {
			contextLogger.Warnf("app: 400 Bad Request - path parameter must be a valid v4 uuid")
			clientError(w, http.StatusBadRequest, ErrCodeBadRequest,
				"path parameter must be a valid v4 uuid") // 400
			return
		}

		coupon, err := a.Service.UpdateCoupon(ctx, couponID, request.Void)
		if err == service.ErrCouponNotFound {
			contextLogger.Infof("app: 404 Not Found - coupon %q not found",
				couponID)
			clientError(w, http.StatusNotFound, ErrCodeCouponNotFound,
				"coupon not found") // 404
			return
		}
		if err != nil {
			contextLogger.Errorf("app: a.Service.UpdateCoupon(ctx, couponID=%q, void=%v) failed: %+v",
				couponID, request.Void, err)
			w.WriteHeader(http.StatusInternalServerError) // 500
			return
		}

		w.WriteHeader(http.StatusOK) // 200
		json.NewEncoder(w).Encode(&coupon)
		contextLogger.Infof("app: 200 OK - returning coupon data")
	}
}
