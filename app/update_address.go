package app

import (
	"encoding/json"
	"net/http"

	service "bitbucket.org/andyfusniakteam/ecom-api-go/service/firebase"
)

// UpdateAddressHandler updates an address by addresss UUID
func (a *App) UpdateAddressHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// a.Service.UpdateAddress()

		w.WriteHeader(http.StatusOK) // 200 OK
		json.NewEncoder(w).Encode(service.User{})
	}
}
