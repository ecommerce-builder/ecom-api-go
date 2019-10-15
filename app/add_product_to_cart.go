package app

import (
	"encoding/json"
	"net/http"

	service "bitbucket.org/andyfusniakteam/ecom-api-go/service/firebase"
	log "github.com/sirupsen/logrus"
)

type addProductToCartRequestBody struct {
	CartID    *string `json:"cart_id"`
	ProductID *string `json:"product_id"`
	Qty       *int    `json:"qty"`
}

func validateAddProductRequestBody(request *addProductToCartRequestBody) (bool, string) {
	// cart_id attribute
	if request.CartID == nil {
		return false, "cart_id attribute must be set"
	}
	if !IsValidUUID(*request.CartID) {
		return false, "cart_id attribute must be a valid v4 UUID"
	}

	// product_id attribute
	if request.ProductID == nil {
		return false, "product_id attribute must be set"
	}
	if !IsValidUUID(*request.ProductID) {
		return false, "product_id attribute must be a valid v4 UUID"
	}

	// qty attribute
	if request.Qty == nil {
		return false, "qty attribute must be set"
	}
	if *request.Qty < 1 {
		return false, "qty attribute must be set to at least 1"
	}

	return true, ""
}

// AddProductToCartHandler creates a handler to add a product to a given cart
func (a *App) AddProductToCartHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		contextLogger := log.WithContext(ctx)
		contextLogger.Info("app: AddProductToCartHandler started")

		request := addProductToCartRequestBody{}
		dec := json.NewDecoder(r.Body)
		dec.DisallowUnknownFields()
		if err := dec.Decode(&request); err != nil {
			clientError(w, http.StatusBadRequest, ErrCodeBadRequest, err.Error())
			return
		}

		ok, message := validateAddProductRequestBody(&request)
		if !ok {
			clientError(w, http.StatusBadRequest, ErrCodeBadRequest, message)
			return
		}

		userID := ctx.Value(ecomUIDKey).(string)
		product, err := a.Service.AddProductToCart(ctx, userID, *request.CartID, *request.ProductID, *request.Qty)
		if err == service.ErrCartNotFound {
			clientError(w, http.StatusNotFound, ErrCodeCartNotFound, "cart not found")
			return
		}
		if err == service.ErrUserNotFound {
			// 409 Conflict
			clientError(w, http.StatusConflict, ErrCodeUserNotFound, "The userID inside the JWT did not match any user in the system")
			return
		}
		if err == service.ErrProductNotFound {
			// 409 Conflict
			clientError(w, http.StatusConflict, ErrCodeProductNotFound, "failed to add product with given id to the cart as the product cannot be found")
			return
		}
		if err == service.ErrDefaultPriceListNotFound {
			// 500 Internal Server Error
			contextLogger.Error("ErrDefaultPriceListNotFound")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		if err == service.ErrCartProductExists {
			// 409 Conflict
			clientError(w, http.StatusConflict, ErrCodeCartProductExists, "cart product already in the cart")
			return
		}
		if err == service.ErrProductHasNoPrices {
			// 409 Conflict
			clientError(w, http.StatusConflict, ErrCodeProductHasNoPrices, "can not add to cart as the product prices have not been set")
			return
		}
		if err != nil {
			contextLogger.Errorf("service AddProductToCart(cartID=%q, productUD=%q, qty=%d) failed with error: %+v", *request.CartID, *request.ProductID, *request.Qty, err)
			w.WriteHeader(http.StatusInternalServerError) // 500 Internal Server Error
			return
		}
		w.WriteHeader(http.StatusCreated) // 201 Created
		json.NewEncoder(w).Encode(&product)
	}
}
