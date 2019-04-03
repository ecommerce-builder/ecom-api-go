package app

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/go-chi/chi"
)

// GetAddressHandler gets an address by its address UUID
func (app *App) GetAddressHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		uuid := chi.URLParam(r, "uuid")

		addr, err := app.Service.GetAddress(r.Context(), uuid)
		if err != nil {
			if app.Service.IsNotExist(err) {
				// if ne, ok := err.(*firebase.ResourceError); ok {
				// 	return
				// }
				return
			}
			fmt.Fprintf(os.Stderr, "failed to get address: %v", err)
			return
		}
		w.WriteHeader(http.StatusOK) // 200 OK
		json.NewEncoder(w).Encode(*addr)
	}
}
