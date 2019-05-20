package app

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
)

// ListAdminsHandler creates a handler that returns a list of administrators.
func (a *App) ListAdminsHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		admins, err := a.Service.ListAdmins(r.Context())
		if err != nil {
			fmt.Fprintf(os.Stderr, "service: GetAdmins(ctx) error: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK) // 200 OK
		json.NewEncoder(w).Encode(admins)
	}
}
