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
			clientError(w, http.StatusBadRequest, ErrCodeBadRequest, "cart_id is not a valid UUID v4")
			return
		}

		userID := ctx.Value(ecomUIDKey).(string)
		cartProducts, err := a.Service.GetCartProducts(ctx, userID, cartID)
		if err == service.ErrCartNotFound {
			w.WriteHeader(http.StatusNotFound)
			clientError(w, http.StatusNotFound, ErrCodeCartNotFound, "cart could not be found")
			return
		}
		if err == service.ErrUserNotFound {
			clientError(w, http.StatusNotFound, ErrCodeUserNotFound, "user for this call could not be found")
			return
		}
		if err == service.ErrDefaultPriceListNotFound {
			clientError(w, http.StatusNotFound, ErrCodePriceListNotFound, "user price list could not be found")
			return
		}
		if err != nil {
			contextLogger.Errorf("app: service GetCartProducts(userID=%q, cartID=%q) error: %v", userID, cartID, err)
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
