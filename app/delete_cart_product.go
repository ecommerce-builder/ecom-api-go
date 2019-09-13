package app

import (
	"net/http"

	service "bitbucket.org/andyfusniakteam/ecom-api-go/service/firebase"
	"github.com/go-chi/chi"
	log "github.com/sirupsen/logrus"
)

type cartResponseBody struct {
	Object string `json:"object"`
	*service.Cart
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
				clientError(w, http.StatusNotFound, ErrCodeCartProductNotFound, "cart product not found")
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
