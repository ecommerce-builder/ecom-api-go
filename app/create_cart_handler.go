package app

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	log "github.com/sirupsen/logrus"
)

// CreateCartHandler create a new shopping cart
func (a *App) CreateCartHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Debug("CreateCartHandler started")

		var cart struct {
			UUID string `json:"uuid"`
		}

		uuid, err := a.Service.CreateCart(r.Context())
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to create cart: %v", err)
			return
		}

		log.Debugf("a.Service.CreateCart() returned %s", *uuid)
		cart.UUID = *uuid
		w.WriteHeader(http.StatusCreated) // 201 Created
		json.NewEncoder(w).Encode(cart)
	}
}
