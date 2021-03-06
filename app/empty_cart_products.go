package app

import (
	"net/http"

	service "bitbucket.org/andyfusniakteam/ecom-api-go/service/firebase"
	log "github.com/sirupsen/logrus"
)

// EmptyCartProductsHandler empties the cart of all products (not including coupons).
func (a *App) EmptyCartProductsHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		contextLogger := log.WithContext(ctx)
		contextLogger.Info("app: EmptyCartProductsHandler started")

		cartID := r.URL.Query().Get("cart_id")
		if cartID == "" {
			clientError(w, http.StatusBadRequest, ErrCodeBadRequest, "query parameter cart_id must be set")
			return
		}

		err := a.Service.EmptyCartProducts(ctx, cartID)
		if err == service.ErrCartNotFound {
			clientError(w, http.StatusNotFound, ErrCodeCartNotFound, "cart not found")
			return
		}
		if err != nil {
			contextLogger.Errorf("app: EmptyCartProducts(ctx, cartID=%q) error: %v", cartID, err)
			w.WriteHeader(http.StatusInternalServerError) // 500 Internal Server Error
			return
		}
		w.Header().Del("Content-Type")
		w.WriteHeader(http.StatusNoContent) // 204 No Content
	}
}
