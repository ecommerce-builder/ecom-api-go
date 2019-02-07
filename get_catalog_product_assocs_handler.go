package app
import (
	"net/http"
)

// GetCatalogProductAssocsHandler creates a handler to return the entire catalog
func (app *App) GetCatalogProductAssocsHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
	}
}
