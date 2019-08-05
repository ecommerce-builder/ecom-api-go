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
		SKU string `json:"sku"`
		Qty int    `json:"qty"`
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
		item, err := a.Service.AddItemToCart(ctx, cartID, o.SKU, o.Qty)
		if err != nil {
			if err == service.ErrCartNotFound {
				w.WriteHeader(http.StatusNotFound) // 404 Not Found
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
			contextLogger.Errorf("service AddItemToCart(cartID=%q, sku=%q, qty=%d) failed with error: %v", cartID, o.SKU, o.Qty, err)
			w.WriteHeader(http.StatusInternalServerError) // 500 Internal Server Error
			return
		}
		w.WriteHeader(http.StatusCreated) // 201 Created
		json.NewEncoder(w).Encode(&item)
	}
}
