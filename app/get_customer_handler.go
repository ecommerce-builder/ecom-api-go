package app

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/go-chi/chi"
)

// GetCustomerHandler
func (a *App) GetCustomerHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cuuid := chi.URLParam(r, "cuuid")
		customer, err := a.Service.GetCustomer(r.Context(), cuuid)
		if err != nil {
			fmt.Fprintf(os.Stderr, "service GetCustomer(%s) error: %v", cuuid, err)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK) // 200 OK
		json.NewEncoder(w).Encode(*customer)
	}
}
