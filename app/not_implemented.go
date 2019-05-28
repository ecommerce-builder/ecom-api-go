package app

import (
	"net/http"
)

// NotImplementedHandler creates a handler function that returns 501 Not Implemented.
func (a *App) NotImplementedHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotImplemented) // 501 Not Implemented
	}
}
