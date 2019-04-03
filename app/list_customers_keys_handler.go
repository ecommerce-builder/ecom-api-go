package app

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/go-chi/chi"
)

// ListCustomersDevKeysHandler get a list of addresses
func (a *App) ListCustomersDevKeysHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		uuid := chi.URLParam(r, "uuid")
		apiKeys, err := a.Service.ListCustomersDevKeys(r.Context(), uuid)
		if err != nil {
			fmt.Fprintf(os.Stderr, "service ListCustomersDevKeys(ctx, %s) error: %v", uuid, err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK) // 200 OK
		json.NewEncoder(w).Encode(apiKeys)
	}
}
