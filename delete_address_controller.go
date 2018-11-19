package app

import (
	"fmt"
	"net/http"
	"os"

	"github.com/gorilla/mux"
)

// DeleteAddressController handler
func (a *App) DeleteAddressController() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		params := mux.Vars(r)

		err := a.Service.DeleteAddress(params["aid"])
		if err != nil {
			fmt.Fprintf(os.Stderr, "service DeleteAddress(%s) error: %v", params["aid"], err)
			return
		}

		w.WriteHeader(http.StatusNoContent) // 204 No Content
	}
}
