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
		uuid := chi.URLParam(r, "uuid")
		err := a.Service.DeleteAddress(r.Context(), uuid)
		if err != nil {
			fmt.Fprintf(os.Stderr, "service DeleteAddress(ctx, %s) error: %v", uuid, err)
			return
		}
		w.Header().Del("Content-Type")
		w.WriteHeader(http.StatusNoContent) // 204 No Content
	}
}
