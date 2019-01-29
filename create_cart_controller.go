package app

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	log "github.com/sirupsen/logrus"
)

// CreateCartController handler
func (a *App) CreateCartController() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Debug("CreateCartController started")

		var cart struct {
			CartUUID string `json:"cart_uuid"`
		}

		uuid, err := a.Service.CreateCart(r.Context())
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to create cart: %v", err)
			return
		}

		log.Debugf("a.Service.CreateCart() returned %s", *uuid)
		cart.CartUUID = *uuid

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated) // 201 Created
		json.NewEncoder(w).Encode(cart)
	}
}
