package app

import (
	"net/http"

	"github.com/gorilla/mux"
)

// DeleteCartItemController handler
func (a *App) DeleteCartItemController() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		params := mux.Vars(r)

		count, _ := a.Service.DeleteCartItem(params["ctid"], params["sku"])
		if count == 0 {
			w.WriteHeader(http.StatusNotFound) // 404 Not Found
			return
		}

		w.WriteHeader(http.StatusNoContent) // 204 No Content
	}
}
