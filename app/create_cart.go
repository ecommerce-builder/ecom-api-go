package app

import (
	"encoding/json"
	"net/http"

	log "github.com/sirupsen/logrus"
)

// CreateCartHandler returns an http.HandlerFunc that creates a new shopping cart.
func (a *App) CreateCartHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		contextLogger := log.WithContext(ctx)
		contextLogger.Info("app: CreateCartHandler called")

		cart, err := a.Service.CreateCart(ctx)
		if err != nil {
			contextLogger.Errorf("app: a.Service.CreateCart(ctx) failed: %+v", err)
			w.WriteHeader(http.StatusInternalServerError) // 500 Internal Server Error
			return
		}
		w.WriteHeader(http.StatusCreated) // 201 Created
		json.NewEncoder(w).Encode(&cart)
	}
}
