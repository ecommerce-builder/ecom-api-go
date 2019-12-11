package app

import (
	"encoding/json"
	"net/http"

	service "bitbucket.org/andyfusniakteam/ecom-api-go/service/firebase"
	log "github.com/sirupsen/logrus"
)

type orderRequestBody struct {
	CartID      *string                         `json:"cart_id"`
	ContactName *string                         `json:"contact_name"`
	Email       *string                         `json:"email"`
	UserID      *string                         `json:"user_id"`
	BillingID   *string                         `json:"billing_id"`
	ShippingID  *string                         `json:"shipping_id"`
	Billing     *service.NewOrderAddressRequest `json:"billing"`
	Shipping    *service.NewOrderAddressRequest `json:"shipping"`
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
			contextLogger.Warnf("app: 400 Bad Request - decode error %v",
				err.Error())
			clientError(w, http.StatusBadRequest, ErrCodeBadRequest,
				err.Error()) // 400
			return
		}
		defer r.Body.Close()

		message, ok := validateOrderRequestBody(&req)
		if !ok {
			contextLogger.Warnf("app: 400 Bad Request - validation %v",
				message)
			clientError(w, http.StatusConflict,
				ErrCodeBadRequest, message) // 400
			return
		}

		var order *service.Order
		if req.UserID == nil {
			order, err = a.Service.PlaceGuestOrder(ctx, *req.CartID, *req.ContactName, *req.Email,
				req.Billing, req.Shipping)
		} else {
			order, err = a.Service.PlaceOrder(ctx, *req.CartID,
				*req.UserID, *req.BillingID, *req.ShippingID)
		}

		if err == service.ErrCartNotFound {
			contextLogger.Warn("app: 404 Not Found - cart not found")
			clientError(w, http.StatusNotFound, ErrCodeCartNotFound,
				"cart not found") // 404
		}
		if err == service.ErrCartEmpty {
			contextLogger.Warn("app: 409 Conflict - cart is empty")
			clientError(w, http.StatusConflict, ErrCodeOrderCartEmpty,
				"The cart id you passed contains no items") // 404
			return
		}
		if err == service.ErrUserNotFound {
			contextLogger.Warn("app: 404 Not Found - user not found")
			clientError(w, http.StatusNotFound, ErrCodeOrderUserNotFound,
				"The user with the given user_id could not be found") // 404
			return
		}
		if err == service.ErrAddressNotFound {
			contextLogger.Warnf("app: 404 Not Found - address not found")
			clientError(w, http.StatusNotFound, ErrCodeAddressNotFound,
				"billing or shipping address not found")
			return
		}
		if err != nil {
			contextLogger.Panicf("app: PlaceOrder(ctx, ...) failed with error: %v", err)
			w.WriteHeader(http.StatusInternalServerError) // 500
			return
		}
		w.WriteHeader(http.StatusCreated) // 201 Created
		json.NewEncoder(w).Encode(&order)
	}
}

func validateOrderRequestBody(req *orderRequestBody) (string, bool) {
	// cart_id
	if req.CartID == nil {
		return "cart_id attribute missing", false
	}

	// user_id
	if req.UserID != nil {
		//
		// registered user order
		//
		if !IsValidUUID(*req.UserID) {
			return "attribute user_id must be a valid v4 uuid", false
		}

		// when the user_id is set, the contact_name, email,
		// billing and shipping need not be. Instead the billing_id
		// and shipping_id must point to a valid address.
		if req.ContactName != nil || req.Email != nil || req.Billing != nil || req.Shipping != nil {
			return "attributes contact_name, email, billing and shipping should not be set when the user_id is set", false
		}
		if req.BillingID == nil {
			return "billing_id attribute must be set", false
		}
		if !IsValidUUID(*req.BillingID) {
			return "billing_id attribute must be a valid v4 uuid", false

		}
		if req.ShippingID == nil {
			return "shipping_id attribute must be set", false
		}
		if !IsValidUUID(*req.ShippingID) {
			return "shipping_id attribute must be a valid v4 uuid", false

		}
	} else {
		//
		// guest order
		//
		if req.ContactName == nil {
			return "contact_name attribute missing", false
		}
		if req.Email == nil {
			return "email attribute missing", false
		}

		// billing
		if req.Billing == nil {
			return "billing attribute missing", false
		}
		if req.Billing.ContactName == nil {
			return "billing.contact_name attribute missing", false
		}
		if req.Billing.Addr1 == nil {
			return "billing.addr1 attribute missing", false
		}
		if req.Billing.City == nil {
			return "billing.city attribute missing", false
		}
		if req.Billing.Postcode == nil {
			return "billing.postcode attribute missing", false
		}
		if req.Billing.CountryCode == nil {
			return "billing.country_code attribute missing", false
		}

		// shipping
		if req.Shipping == nil {
			return "shipping attribute missing", false
		}
		if req.Shipping.ContactName == nil {
			return "shipping.contact_name attribute missing", false
		}
		if req.Shipping.Addr1 == nil {
			return "shipping.addr1 attribute missing", false
		}
		if req.Shipping.City == nil {
			return "shipping.city attribute missing", false
		}
		if req.Shipping.Postcode == nil {
			return "shipping.postcode attribute missing", false
		}
		if req.Shipping.CountryCode == nil {
			return "shipping.country_code attribute missing", false
		}
	}

	return "", true
}
