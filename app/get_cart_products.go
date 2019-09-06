package app

import (
	"encoding/json"
	"net/http"

	service "bitbucket.org/andyfusniakteam/ecom-api-go/service/firebase"
	log "github.com/sirupsen/logrus"
)

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
