package app

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/go-chi/chi"
)

// GetAddressController handler
func (a *App) GetAddressController() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		auuid := chi.URLParam(r, "auuid")

		a, err := a.Service.GetAddress(auuid)
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to get address: %v", err)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK) // 200 OK
		json.NewEncoder(w).Encode(*a)
	}
}
