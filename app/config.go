package app

import (
	"encoding/json"
	"net/http"
)

type configResponseBody struct {
	Object   string             `json:"object"`
	Firebase *FirebaseSystemEnv `json:"firebaseConfig"`
}

// ConfigHandler returns a handler function that returns API configuration
// including Firebase public key.
func (a *App) ConfigHandler(se FirebaseSystemEnv) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK) // 200 OK
		res := configResponseBody{
			Object:   "config",
			Firebase: &se,
		}
		json.NewEncoder(w).Encode(res)
	}
}
