package app

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi"
)

// ListAddressesController handler
func (a *App) ListAddressesController() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cuuid := chi.URLParam(r, "cuuid")
		addresses, _ := a.Service.GetAddresses(cuuid)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK) // 200 OK
		json.NewEncoder(w).Encode(addresses)
	}
}
