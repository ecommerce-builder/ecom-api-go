package app

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"bitbucket.org/andyfusniakteam/ecom-api-go/service/firebase"
)

// GetCatalogAssocsHandler creates a handler to return the entire catalog
func (app *App) GetCatalogAssocsHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cpo, err := app.Service.GetCatalogAssocs(r.Context())
		if err != nil {
			fmt.Fprintf(os.Stderr, "service GetCatalogAssocs(ctx) error: %v", err)
			w.WriteHeader(http.StatusInternalServerError) // 500
			json.NewEncoder(w).Encode(struct {
				Status  int    `json:"status"`
				Code    string `json:"code"`
				Message string `json:"message"`
			}{
				http.StatusInternalServerError,
				ErrCodeInternalServerError,
				err.Error(),
			})
			return
		}
		as := make([]*firebase.Assoc, 0)
		for _, v := range cpo {
			as = append(as, v)
		}
		w.WriteHeader(http.StatusOK) // 200 OK
		json.NewEncoder(w).Encode(as)
	}
}
