package app

import (
	"encoding/json"
	"net/http"

	service "bitbucket.org/andyfusniakteam/ecom-api-go/service/firebase"
	"github.com/go-chi/chi"
	log "github.com/sirupsen/logrus"
)

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
			clientError(w, http.StatusBadRequest, ErrCodeBadRequest, err.Error())
			return
		}

		userID := ctx.Value(ecomUIDKey).(string)
		product, err := a.Service.UpdateCartProduct(ctx, userID, cartProductID, request.Qty)
		if err != nil {
			if err == service.ErrCartProductNotFound {
				contextLogger.Debugf("app: Cart Product (cartProductID=%q) not found", cartProductID)
				clientError(w, http.StatusNotFound, ErrCodeCartProductNotFound, "cart product not found")
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
