package app

import (
	"encoding/json"
	"fmt"
	"net/http"

	service "bitbucket.org/andyfusniakteam/ecom-api-go/service/firebase"
	log "github.com/sirupsen/logrus"
)

type orderRequestBody struct {
	ContactName string              `json:"contact_name"`
	CartID      *string             `json:"cart_id"`
	Email       string              `json:"email"`
	Billing     *service.NewAddress `json:"billing_address"`
	Shipping    *service.NewAddress `json:"shipping_address"`
}

type orderResponseBody struct {
	Object string `json:"object"`
	*service.Order
}

func validateOrderRequestBody(req *orderRequestBody) ([]string, bool) {
	var errs []string
	if req.CartID == nil {
		errs = append(errs, "cart_id missing")
	}

	if len(errs) > 0 {
		return errs, false
	}
	return nil, true
}

// PlaceOrderHandler returns an HTTP handler that places a new order.
func (a *App) PlaceOrderHandler() http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		contextLogger := log.WithContext(ctx)
		contextLogger.Info("App: PlaceOrderHandler started")

		req := orderRequestBody{}
		err := json.NewDecoder(r.Body).Decode(&req)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest) // 400 Bad Request
			json.NewEncoder(w).Encode(struct {
				Status  int    `json:"status"`
				Code    string `json:"code"`
				Message string `json:"message"`
			}{
				http.StatusBadRequest,
				ErrCodeBadRequest,
				err.Error(),
			})
			return
		}
		defer r.Body.Close()

		valErrs, ok := validateOrderRequestBody(&req)
		if !ok {
			fmt.Println(valErrs)
			w.WriteHeader(http.StatusConflict) // 409 Conflict
			return
		}

		fmt.Printf("%#v\n", req)

		order, err := a.Service.PlaceOrder(ctx, req.ContactName, req.Email, req.Billing, req.Shipping)
		if err != nil {
			contextLogger.Panicf("App: PlaceOrder(ctx, %q, %q, ...) failed with error: %v", req.ContactName, req.Email, err)
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
