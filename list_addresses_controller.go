package app

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
)

// ListAddressesController handler
func (a *App) ListAddressesController() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		params := mux.Vars(r)

		addresses, _ := a.Service.GetAddresses(params["cid"])

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK) // 200 OK
		json.NewEncoder(w).Encode(addresses)
	}
}
