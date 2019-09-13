package app

import (
	"encoding/json"
	"net/http"

	service "bitbucket.org/andyfusniakteam/ecom-api-go/service/firebase"
	log "github.com/sirupsen/logrus"
)

type orderRequestBody struct {
	CartID      *string             `json:"cart_id"`
	ContactName *string             `json:"contact_name"`
	Email       *string             `json:"email"`
	CustomerID  *string             `json:"customer_id"`
	Billing     *service.NewAddress `json:"billing_address"`
	Shipping    *service.NewAddress `json:"shipping_address"`
}

type orderResponseBody struct {
	Object string `json:"object"`
	*service.Order
}

func validateOrderRequestBody(req *orderRequestBody) (string, bool) {
	if req.CartID == nil {
		return "cart_id missing", false
	}

	if req.CustomerID != nil {
		if req.ContactName != nil || req.Email != nil {
			return "Either set customer_id for customer orders or set contact_name with email for guest orders - not both", false
		}
	} else {
		if req.ContactName == nil || req.Email == nil {
			return "For placing guest orders set both contact_name and email", false
		}
	}
	return "", true
}

// PlaceOrderHandler returns an HTTP handler that places a new order.
func (a *App) PlaceOrderHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		contextLogger := log.WithContext(ctx)
		contextLogger.Info("app: PlaceOrderHandler started")

		req := orderRequestBody{}
		dec := json.NewDecoder(r.Body)
		dec.DisallowUnknownFields()
		err := dec.Decode(&req)
		if err != nil {
			clientError(w, http.StatusBadRequest, ErrCodeBadRequest, err.Error())
			return
		}
		defer r.Body.Close()

		message, ok := validateOrderRequestBody(&req)
		if !ok {
			clientError(w, http.StatusConflict, ErrCodeBadRequest, message)
			return
		}

		contextLogger.Debugf("app: a.Service.PlaceOrder(ctx, req.ContextName=%v, req.Email=%v, res.CustomerID=%v, %q, ...)", req.ContactName, req.Email, req.CustomerID, *req.CartID)
		order, err := a.Service.PlaceOrder(ctx, req.ContactName, req.Email, req.CustomerID, *req.CartID, req.Billing, req.Shipping)
		if err != nil {
			if err == service.ErrCartEmpty {
				clientError(w, http.StatusNotFound, ErrCodeOrderCartEmpty, "The cart id you passed contains no items")
				return
			} else if err == service.ErrUserNotFound {
				clientError(w, http.StatusNotFound, ErrCodeOrderUserNotFound, "The user with the given user_id could not be found")
				return
			}
			contextLogger.Panicf("app: PlaceOrder(ctx, %q, %q, ...) failed with error: %v", *req.ContactName, *req.Email, err)
			w.WriteHeader(http.StatusInternalServerError) // 500 Internal Server Error
			return
		}
		res := orderResponseBody{
			Object: "order",
			Order:  order,
		}
		w.WriteHeader(http.StatusCreated) // 201 Created
		json.NewEncoder(w).Encode(res)
	}
}
