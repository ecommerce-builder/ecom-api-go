package app

import (
	"encoding/json"
	"net/http"

	service "bitbucket.org/andyfusniakteam/ecom-api-go/service/firebase"
	"github.com/go-chi/chi"

	log "github.com/sirupsen/logrus"
)

// UpdateCartItemHandler creates a handler to add an item to a given cart
func (app *App) UpdateCartItemHandler() http.HandlerFunc {
	type qtyRequestBody struct {
		Qty int `json:"qty"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		contextLogger := log.WithContext(ctx)
		contextLogger.Info("App: UpdateCartItemHandler started")

		id := chi.URLParam(r, "id")
		sku := chi.URLParam(r, "sku")
		o := qtyRequestBody{}
		err := json.NewDecoder(r.Body).Decode(&o)
		if err != nil {
			http.Error(w, err.Error(), 400)
			return
		}
		item, err := app.Service.UpdateCartItem(ctx, id, sku, o.Qty)
		if err != nil {
			if err == service.ErrCartItemNotFound {
				w.WriteHeader(http.StatusNotFound)
				return
			}
			contextLogger.Errorf("service UpdateCartItem(ctx, %q, %q, %d) error: %v", id, sku, o.Qty, err)
			w.WriteHeader(http.StatusInternalServerError) // 500 Internal Server Error
			return
		}
		res := cartItemResponseBody{
			Object:   "cart_item",
			CartItem: item,
		}
		w.WriteHeader(http.StatusOK) // 200 OK
		json.NewEncoder(w).Encode(res)
	}
}
