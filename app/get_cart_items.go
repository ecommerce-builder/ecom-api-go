package app

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi"
	log "github.com/sirupsen/logrus"
)

// GetCartItemsHandler returns a list of all cart items
func (a *App) GetCartItemsHandler() http.HandlerFunc {
	type itemsListResponseBody struct {
		Object string                  `json:"object"`
		Items  []*cartItemResponseBody `json:"items"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		contextLogger := log.WithContext(ctx)
		contextLogger.Info("App: GetCartItemsHandler started")

		id := chi.URLParam(r, "id")
		cartItems, err := a.Service.GetCartItems(ctx, id)
		if err != nil {
			contextLogger.Errorf("service GetCartItems(%s) error: %v", id, err)
			w.WriteHeader(http.StatusInternalServerError) // 500 Internal Server Error
			return
		}

		cartResponseItems := make([]*cartItemResponseBody, 0, len(cartItems))
		for _, t := range cartItems {
			rt := cartItemResponseBody{
				Object:   "cart_item",
				CartItem: t,
			}
			cartResponseItems = append(cartResponseItems, &rt)
		}

		res := itemsListResponseBody{
			Object: "list",
			Items:  cartResponseItems,
		}

		w.WriteHeader(http.StatusOK) // 200 OK
		json.NewEncoder(w).Encode(&res)
	}
}
