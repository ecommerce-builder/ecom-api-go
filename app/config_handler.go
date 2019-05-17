package app

import (
	"encoding/json"
	"net/http"
)

// ConfigHandler returns a handler function that returns API configuration
// including Firebase public key.
func (a *App) ConfigHandler(se GoogSystemEnv) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK) // 200 OK
		json.NewEncoder(w).Encode(se)
	}
}
