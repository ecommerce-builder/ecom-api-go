package app

import (
	"fmt"
	"net/http"
	"os"

	"github.com/go-chi/chi"
)

// DeleteAddressHandler deletes an address record
func (a *App) DeleteAddressHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		auuid := chi.URLParam(r, "auuid")

		err := a.Service.DeleteAddress(r.Context(), auuid)
		if err != nil {
			fmt.Fprintf(os.Stderr, "service DeleteAddress(ctx, %s) error: %v", auuid, err)
			return
		}

		w.WriteHeader(http.StatusNoContent) // 204 No Content
	}
}
