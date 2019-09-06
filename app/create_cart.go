package app

import (
	"encoding/json"
	"net/http"

	log "github.com/sirupsen/logrus"
)

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
