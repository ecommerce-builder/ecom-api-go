package app

import (
	"encoding/json"
	"net/http"

	service "bitbucket.org/andyfusniakteam/ecom-api-go/service/firebase"
	log "github.com/sirupsen/logrus"
)

// ListCouponsHandler returns a http.HandlerFunc that returns
// a list of coupons.
func (a *App) ListCouponsHandler() http.HandlerFunc {
	type response struct {
		Object string            `json:"object"`
		Data   []*service.Coupon `json:"data"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		contextLogger := log.WithContext(ctx)
		contextLogger.Info("app: ListCouponsHandler called")

		coupons, err := a.Service.GetCoupons(ctx)
		if err != nil {
			contextLogger.Errorf("app: 500 Internal Server Error - a.Service.GetAllInventory(ctx): %+v", err)
			w.WriteHeader(http.StatusInternalServerError) // 500
			return
		}

		list := response{
			Object: "list",
			Data:   coupons,
		}
		w.WriteHeader(http.StatusOK) // 200
		json.NewEncoder(w).Encode(&list)
		contextLogger.Infof("app: 200 OK - returning coupons")
	}
}
