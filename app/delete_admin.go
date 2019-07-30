package app

import (
	"net/http"

	"github.com/go-chi/chi"
	log "github.com/sirupsen/logrus"
)

// DeleteAdminHandler creates an HTTP handler that deletes an administrator.
func (a *App) DeleteAdminHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		contextLogger := log.WithContext(ctx)
		contextLogger.Info("App: DeleteAdminHandler started")

		id := chi.URLParam(r, "id")
		if err := a.Service.DeleteAdmin(ctx, id); err != nil {
			contextLogger.Errorf("service DeleteAdmin(ctx, %s) failed with error: %v", id, err)
			w.WriteHeader(http.StatusInternalServerError) // 500 Internal Server Error
			return
		}
		w.Header().Del("Content-Type")
		w.WriteHeader(http.StatusNoContent) // 204 No Content
	}
}
