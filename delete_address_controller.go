package app

import (
	"fmt"
	"net/http"
	"os"

	"github.com/go-chi/chi"
)

// DeleteAddressController handler
func (a *App) DeleteAddressController() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		auuid := chi.URLParam(r, "auuid")

		err := a.Service.DeleteAddress(auuid)
		if err != nil {
			fmt.Fprintf(os.Stderr, "service DeleteAddress(%s) error: %v", auuid, err)
			return
		}

		w.WriteHeader(http.StatusNoContent) // 204 No Content
	}
}
