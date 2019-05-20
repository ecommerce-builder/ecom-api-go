package app

import (
	"fmt"
	"net/http"
	"os"

	"github.com/go-chi/chi"
)

// DeleteAdminHandler creates an HTTP handler that deletes an administrator. 
func (a *App) DeleteAdminHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		uuid := chi.URLParam(r, "uuid")
		err := a.Service.DeleteAdmin(r.Context(), uuid)
		if err != nil {
			fmt.Fprintf(os.Stderr, "service DeleteAdmin(ctx, %s) error: %v", uuid, err)
			return
		}
		w.Header().Del("Content-Type")
		w.WriteHeader(http.StatusNoContent) // 204 No Content
	}
}
