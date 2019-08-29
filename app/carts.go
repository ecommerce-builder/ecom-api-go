package app

import (
	"encoding/json"
	"net/http"
	"regexp"

	service "bitbucket.org/andyfusniakteam/ecom-api-go/service/firebase"
	"github.com/go-chi/chi"
	log "github.com/sirupsen/logrus"
)

// AddProductToCartHandler creates a handler to add a product to a given cart
func (a *App) AddProductToCartHandler() http.HandlerFunc {
	type requestBody struct {
		CartID    string `json:"cart_id"`
		ProductID string `json:"product_id"`
		Qty       int    `json:"qty"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		contextLogger := log.WithContext(ctx)
		contextLogger.Info("app: AddProductToCartHandler started")

		request := requestBody{}
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(struct {
				Status  int    `json:"status"`
				Code    string `json:"code"`
				Message string `json:"message"`
			}{
				http.StatusBadRequest,
				ErrCodeBadRequest,
				"bad request",
			})
			return
		}

		product, err := a.Service.AddProductToCart(ctx, request.CartID, request.ProductID, request.Qty)
		if err != nil {
			if err == service.ErrCartNotFound {
				w.WriteHeader(http.StatusNotFound) // 404 Not Found
				return
			} else if err == service.ErrUserNotFound {
				w.WriteHeader(http.StatusConflict) // 409 Conflict
				json.NewEncoder(w).Encode(struct {
					Status  int    `json:"status"`
					Code    string `json:"code"`
					Message string `json:"message"`
				}{
					http.StatusConflict,
					ErrCodeUserNotFound,
					"The userID inside the JWT did not match any user in the system",
				})
				return
			} else if err == service.ErrProductNotFound {
				w.WriteHeader(http.StatusConflict) // 409 Conflict
				json.NewEncoder(w).Encode(struct {
					Status  int    `json:"status"`
					Code    string `json:"code"`
					Message string `json:"message"`
				}{
					http.StatusConflict,
					ErrCodeProductNotFound,
					"failed to add product with given id to the cart as the product cannot be found",
				})
				return
			} else if err == service.ErrDefaultPriceListNotFound {
				contextLogger.Error("ErrDefaultPriceListNotFound")
				w.WriteHeader(http.StatusInternalServerError) // 500 Internal Server Error
				return
			} else if err == service.ErrCartProductExists {
				w.WriteHeader(http.StatusConflict) // 409 Conflict
				json.NewEncoder(w).Encode(struct {
					Status  int    `json:"status"`
					Code    string `json:"code"`
					Message string `json:"message"`
				}{
					http.StatusConflict,
					ErrCodeCartProductExists,
					"cart product already in the cart",
				})
				return
			}

			contextLogger.Errorf("service AddProductToCart(cartID=%q, productUD=%q, qty=%d) failed with error: %v", request.CartID, request.ProductID, request.Qty, err)
			w.WriteHeader(http.StatusInternalServerError) // 500 Internal Server Error
			return
		}
		w.WriteHeader(http.StatusCreated) // 201 Created
		json.NewEncoder(w).Encode(&product)
	}
}

type cartResponseBody struct {
	Object string `json:"object"`
	*service.Cart
}

// CreateCartHandler create a new shopping cart
func (a *App) CreateCartHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		contextLogger := log.WithContext(ctx)

		cart, err := a.Service.CreateCart(ctx)
		if err != nil {
			contextLogger.Errorf("app: failed to create cart: %v", err)
			w.WriteHeader(http.StatusInternalServerError) // 500 Internal Server Error
			return
		}

		response := cartResponseBody{
			Object: "cart",
			Cart:   cart,
		}
		w.WriteHeader(http.StatusCreated) // 201 Created
		json.NewEncoder(w).Encode(&response)
	}
}

// GetCartProductsHandler returns a list of all cart products.
func (a *App) GetCartProductsHandler() http.HandlerFunc {
	type responseBody struct {
		Object string                 `json:"object"`
		Data   []*service.CartProduct `json:"data"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		contextLogger := log.WithContext(ctx)
		contextLogger.Info("app: GetCartProductsHandler started")

		cartID := r.URL.Query().Get("cart_id")
		if IsValidUUID(cartID) == false {
			w.WriteHeader(http.StatusBadRequest) // 400 Bad Request
			json.NewEncoder(w).Encode(struct {
				Status  int    `json:"status"`
				Code    string `json:"code"`
				Message string `json:"message"`
			}{
				http.StatusBadRequest,
				ErrCodeBadRequest,
				"cart_id is not a valid UUID v4",
			})
			return
		}

		cartProducts, err := a.Service.GetCartProducts(ctx, cartID)
		if err != nil {
			if err == service.ErrCartNotFound {
				w.WriteHeader(http.StatusNotFound)
				return
			}
			contextLogger.Errorf("service GetCartProducts(cartID=%q) error: %v", cartID, err)
			w.WriteHeader(http.StatusInternalServerError) // 500 Internal Server Error
			return
		}
		response := responseBody{
			Object: "list",
			Data:   cartProducts,
		}
		w.WriteHeader(http.StatusOK) // 200 OK
		json.NewEncoder(w).Encode(&response)
	}
}

// IsValidUUID checks for a valid UUID v4.
func IsValidUUID(uuid string) bool {
	r := regexp.MustCompile("^[a-fA-F0-9]{8}-[a-fA-F0-9]{4}-4[a-fA-F0-9]{3}-[8|9|aA|bB][a-fA-F0-9]{3}-[a-fA-F0-9]{12}$")
	return r.MatchString(uuid)
}

// UpdateCartProductHandler creates a handler to update a cart product.
func (a *App) UpdateCartProductHandler() http.HandlerFunc {
	type requestBody struct {
		Qty int `json:"qty"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		contextLogger := log.WithContext(ctx)
		contextLogger.Info("app: UpdateCartProductHandler started")

		cartProductID := chi.URLParam(r, "id")

		request := requestBody{}
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(struct {
				Status  int    `json:"status"`
				Code    string `json:"code"`
				Message string `json:"message"`
			}{
				http.StatusBadRequest,
				ErrCodeBadRequest,
				"bad request",
			})
			return
		}

		product, err := a.Service.UpdateCartProduct(ctx, cartProductID, request.Qty)
		if err != nil {
			if err == service.ErrCartProductNotFound {
				contextLogger.Debugf("app: Cart Product (cartProductID=%q) not found", cartProductID)
				w.WriteHeader(http.StatusNotFound) // 404 Not Found
				json.NewEncoder(w).Encode(struct {
					Status  int    `json:"status"`
					Code    string `json:"code"`
					Message string `json:"message"`
				}{
					http.StatusNotFound,
					ErrCodeCartProductNotFound,
					"cart product not found",
				})
				return
			}
			contextLogger.Errorf("app: a.Service.UpdateCartProduct(ctx, cartProductID=%q, request.Qty=%d) error: %v", cartProductID, request.Qty, err)
			w.WriteHeader(http.StatusInternalServerError) // 500 Internal Server Error
			return
		}
		w.WriteHeader(http.StatusOK) // 200 OK
		json.NewEncoder(w).Encode(&product)
	}
}

// DeleteCartProductHandler creates a handler to delete a cart product.
func (a *App) DeleteCartProductHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		contextLogger := log.WithContext(ctx)
		contextLogger.Info("app: DeleteCartProductHandler started")

		cartProductID := chi.URLParam(r, "id")
		if err := a.Service.DeleteCartProduct(ctx, cartProductID); err != nil {
			if err == service.ErrCartProductNotFound {
				contextLogger.Debugf("app: CartProduct (cartProductID=%q) not found", cartProductID)
				w.WriteHeader(http.StatusNotFound) // 404 Not Found
				json.NewEncoder(w).Encode(struct {
					Status  int    `json:"status"`
					Code    string `json:"code"`
					Message string `json:"message"`
				}{
					http.StatusNotFound,
					ErrCodeCartProductNotFound,
					"cart product not found",
				})
				return
			}
			contextLogger.Errorf("app: a.Service.DeleteCartProduct(ctx, cartProductID=%q) error: %v", cartProductID, err)
			w.WriteHeader(http.StatusInternalServerError) // 500 Internal Server Error
			return
		}

		w.Header().Del("Content-Type")
		w.WriteHeader(http.StatusNoContent) // 204 No Content
	}
}

// EmptyCartProductsHandler empties the cart of all products (not including coupons).
func (a *App) EmptyCartProductsHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		contextLogger := log.WithContext(ctx)
		contextLogger.Info("app: EmptyCartProductsHandler started")

		cartID := r.URL.Query().Get("cart_id")

		if err := a.Service.EmptyCartProducts(ctx, cartID); err != nil {
			if err == service.ErrCartNotFound {
				w.WriteHeader(http.StatusNotFound)
				return
			} else if err == service.ErrCartContainsNoProducts {
				w.WriteHeader(http.StatusConflict) // 409 Conflict
				json.NewEncoder(w).Encode(struct {
					Status  int    `json:"status"`
					Code    string `json:"code"`
					Message string `json:"message"`
				}{
					http.StatusConflict,
					ErrCodeCartContainsNoProducts,
					"OpEmptyCartProducts cannot be called if a cart contains no products",
				})
				return
			}
			contextLogger.Errorf("service EmptyCartProducts(ctx, cartID=%q) error: %v", cartID, err)
			w.WriteHeader(http.StatusInternalServerError) // 500 Internal Server Error
			return
		}
		w.Header().Del("Content-Type")
		w.WriteHeader(http.StatusNoContent) // 204 No Content
	}
}
