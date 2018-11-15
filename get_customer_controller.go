package app

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/gorilla/mux"
)

// GetCustomerController handler
func (a *App) GetCustomerController() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		params := mux.Vars(r)

		customer, err := a.Service.GetCustomer(params["cid"])
		if err != nil {
			fmt.Fprintf(os.Stderr, "service GetCustomer(%s) error: %v", params["cid"], err)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK) // 200 OK
		json.NewEncoder(w).Encode(customer)
	}
}
