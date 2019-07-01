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

		uuid := chi.URLParam(r, "uuid")
		addr, err := app.Service.GetAddress(ctx, uuid)
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
		res := addressResponseBody{
			Object:  "address",
			Address: addr,
		}
		w.WriteHeader(http.StatusOK) // 200 OK
		json.NewEncoder(w).Encode(res)
	}
}
