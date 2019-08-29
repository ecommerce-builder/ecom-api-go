package app

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi"
	log "github.com/sirupsen/logrus"
)

// GetAddressHandler gets an address by its address UUID
func (app *App) GetAddressHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		contextLogger := log.WithContext(ctx)
		contextLogger.Info("App: GetAddressHandler started")

		id := chi.URLParam(r, "id")
		addr, err := app.Service.GetAddress(ctx, id)
		if err != nil {
			if app.Service.IsNotExist(err) {
				// if ne, ok := err.(*firebase.ResourceError); ok {
				// 	return
				// }
				return
			}
			contextLogger.Errorf("failed to get address: %v", err)
			return
		}

		response := addressResponseBody{
			Object:  "address",
			Address: addr,
		}
		w.WriteHeader(http.StatusOK) // 200 OK
		json.NewEncoder(w).Encode(&response)
	}
}
