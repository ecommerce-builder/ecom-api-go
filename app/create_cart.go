package app

import (
	"encoding/json"
	"net/http"

	log "github.com/sirupsen/logrus"
)

type cartResponseBody struct {
	Object string `json:"object"`
	ID     string `json:"id"`
}

// CreateCartHandler create a new shopping cart
func (a *App) CreateCartHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		contextLogger := log.WithContext(ctx)

		id, err := a.Service.CreateCart(ctx)
		if err != nil {
			contextLogger.Errorf("failed to create cart: %v", err)
			w.WriteHeader(http.StatusInternalServerError) // 500 Internal Server Error
			return
		}

		res := cartResponseBody{
			Object: "cart",
			ID:     *id,
		}
		w.WriteHeader(http.StatusCreated) // 201 Created
		json.NewEncoder(w).Encode(res)
	}
}
