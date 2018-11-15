package app

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/gorilla/mux"
)

// GetAddressController handler
func (a *App) GetAddressController() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		params := mux.Vars(r)

		a, err := a.Service.GetAddress(params["aid"])
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to create cart: %v", err)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK) // 200 OK
		json.NewEncoder(w).Encode(a)
	}
}
