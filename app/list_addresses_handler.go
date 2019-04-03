package app

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi"
)

// ListAddressesHandler get a list of addresses
func (a *App) ListAddressesHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cuuid := chi.URLParam(r, "cuuid")
		addresses, _ := a.Service.GetAddresses(r.Context(), cuuid)
		w.WriteHeader(http.StatusOK) // 200 OK
		json.NewEncoder(w).Encode(addresses)
	}
}
