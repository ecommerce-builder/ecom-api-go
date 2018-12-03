package app

import (
	"encoding/json"
	"net/http"
)

// UpdateAddressController handler
func (a *App) UpdateAddressController() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// a.Service.UpdateAddress()

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNoContent) // 204 No Content
		json.NewEncoder(w).Encode(Customer{})
	}
}
