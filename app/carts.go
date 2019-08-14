package app

import (
	"encoding/json"
	"net/http"

	service "bitbucket.org/andyfusniakteam/ecom-api-go/service/firebase"
	"github.com/go-chi/chi"
	log "github.com/sirupsen/logrus"
)

// AddItemToCartHandler creates a handler to add an item to a given cart
func (a *App) AddItemToCartHandler() http.HandlerFunc {
	type itemRequestBody struct {
		ProductID string `json:"product_id"`
		Qty       int    `json:"qty"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		contextLogger := log.WithContext(ctx)
		contextLogger.Info("App: AddItemToCartHandler started")

		cartID := chi.URLParam(r, "cart_id")
		o := itemRequestBody{}
		if err := json.NewDecoder(r.Body).Decode(&o); err != nil {
			http.Error(w, err.Error(), 400)
			return
		}
		item, err := a.Service.AddItemToCart(ctx, cartID, o.ProductID, o.Qty)
		if err != nil {
			if err == service.ErrCartNotFound {
				w.WriteHeader(http.StatusNotFound) // 404 Not Found
				return
			} else if err == service.ErrCustomerNotFound {
				w.WriteHeader(http.StatusConflict) // 409 Conflict
				json.NewEncoder(w).Encode(struct {
					Status  int    `json:"status"`
					Code    string `json:"code"`
					Message string `json:"message"`
				}{
					http.StatusConflict,
					ErrCodeCustomerNotFound,
					"The customerID inside the JWT did not match any customers in the system",
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
			} else if err == service.ErrDefaultPricingTierMissing {
				contextLogger.Error("ErrDefaultPricingTierMissing")
				w.WriteHeader(http.StatusInternalServerError) // 500 Internal Server Error
				return
			} else if err == service.ErrCartItemAlreadyExists {
				w.WriteHeader(http.StatusConflict) // 409 Conflict
				json.NewEncoder(w).Encode(struct {
					Status  int    `json:"status"`
					Code    string `json:"code"`
					Message string `json:"message"`
				}{
					http.StatusConflict,
					ErrCodeCartAlreadyExists,
					"cart item already exists in the cart",
				})
				return
			}

			contextLogger.Errorf("service AddItemToCart(cartID=%q, productUD=%q, qty=%d) failed with error: %v", cartID, o.ProductID, o.Qty, err)
			w.WriteHeader(http.StatusInternalServerError) // 500 Internal Server Error
			return
		}
		w.WriteHeader(http.StatusCreated) // 201 Created
		json.NewEncoder(w).Encode(&item)
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
			contextLogger.Errorf("failed to create cart: %v", err)
			w.WriteHeader(http.StatusInternalServerError) // 500 Internal Server Error
			return
		}

		res := cartResponseBody{
			Object: "cart",
			Cart:   cart,
		}
		w.WriteHeader(http.StatusCreated) // 201 Created
		json.NewEncoder(w).Encode(res)
	}
}

// GetCartItemsHandler returns a list of all cart items
func (a *App) GetCartItemsHandler() http.HandlerFunc {
	type itemsListResponseBody struct {
		Object string              `json:"object"`
		Data   []*service.CartItem `json:"data"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		contextLogger := log.WithContext(ctx)
		contextLogger.Info("App: GetCartItemsHandler started")

		cartID := chi.URLParam(r, "cart_id")
		cartItems, err := a.Service.GetCartItems(ctx, cartID)
		if err != nil {
			if err == service.ErrCartNotFound {
				w.WriteHeader(http.StatusNotFound)
				return
			}
			contextLogger.Errorf("service GetCartItems(cartID=%q) error: %v", cartID, err)
			w.WriteHeader(http.StatusInternalServerError) // 500 Internal Server Error
			return
		}
		res := itemsListResponseBody{
			Object: "list",
			Data:   cartItems,
		}
		w.WriteHeader(http.StatusOK) // 200 OK
		json.NewEncoder(w).Encode(&res)
	}
}

// UpdateCartItemHandler creates a handler to add an item to a given cart
func (app *App) UpdateCartItemHandler() http.HandlerFunc {
	type qtyRequestBody struct {
		Qty int `json:"qty"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		contextLogger := log.WithContext(ctx)
		contextLogger.Info("App: UpdateCartItemHandler started")

		cartID := chi.URLParam(r, "cart_id")
		sku := chi.URLParam(r, "sku")
		o := qtyRequestBody{}
		err := json.NewDecoder(r.Body).Decode(&o)
		if err != nil {
			http.Error(w, err.Error(), 400)
			return
		}
		item, err := app.Service.UpdateCartItem(ctx, cartID, sku, o.Qty)
		if err != nil {
			if err == service.ErrCartItemNotFound {
				w.WriteHeader(http.StatusNotFound)
				return
			}
			contextLogger.Errorf("service UpdateCartItem(ctx, cartID=%q, sku=%q, qty=%d) error: %v", cartID, sku, o.Qty, err)
			w.WriteHeader(http.StatusInternalServerError) // 500 Internal Server Error
			return
		}
		w.WriteHeader(http.StatusOK) // 200 OK
		json.NewEncoder(w).Encode(&item)
	}
}

// DeleteCartItemHandler creates a handler to delete an item from the cart with the given cart UUID.
func (a *App) DeleteCartItemHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		contextLogger := log.WithContext(ctx)
		contextLogger.Info("App: DeleteCartItemHandler started")

		cartID := chi.URLParam(r, "cart_id")
		sku := chi.URLParam(r, "sku")
		count, err := a.Service.DeleteCartItem(ctx, cartID, sku)
		if err != nil {
			if err == service.ErrCartNotFound {
				w.WriteHeader(http.StatusNotFound)
				return
			}
			contextLogger.Errorf("service DeleteCartItem(ctx, cartID=%q, sku=%q) error: %v", cartID, sku, err)
			w.WriteHeader(http.StatusInternalServerError) // 500 Internal Server Error
			return
		}
		if count == 0 {
			w.WriteHeader(http.StatusNotFound) // 404 Not Found
			return
		}
		if err != nil {
			contextLogger.Errorf("DeleteCartItem(ctx, cartID=%q, sku=%q) failed: %v", cartID, sku, err)
			w.WriteHeader(http.StatusInternalServerError) // 500 Internal Server Error
			return
		}
		w.Header().Del("Content-Type")
		w.WriteHeader(http.StatusNoContent) // 204 No Content
	}
}

// EmptyCartItemsHandler empties the cart of all items. This does not remove coupons.
func (a *App) EmptyCartItemsHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		contextLogger := log.WithContext(ctx)
		contextLogger.Info("App: EmptyCartItemsHandler started")

		cartID := chi.URLParam(r, "cart_id")
		if err := a.Service.EmptyCartItems(ctx, cartID); err != nil {
			if err == service.ErrCartNotFound {
				w.WriteHeader(http.StatusNotFound)
				return
			} else if err == service.ErrCartContainsNoItems {
				w.WriteHeader(http.StatusConflict) // 409 Conflict
				json.NewEncoder(w).Encode(struct {
					Status  int    `json:"status"`
					Code    string `json:"code"`
					Message string `json:"message"`
				}{
					http.StatusConflict,
					ErrCodeCartContainsNoItems,
					"OpEmptyCartItems cannot be called if a cart contains no items",
				})
				return
			}
			contextLogger.Errorf("service EmptyCartItems(ctx, cartID=%q) error: %v", cartID, err)
			w.WriteHeader(http.StatusInternalServerError) // 500 Internal Server Error
			return
		}
		w.Header().Del("Content-Type")
		w.WriteHeader(http.StatusNoContent) // 204 No Content
	}
}
