package app

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/go-chi/chi"
)

// GenerateCustomerDevKeyHandler creates a new API Key for a given customer
func (a *App) GenerateCustomerDevKeyHandler() http.HandlerFunc {
	type apiDevResponse struct {
		UUID     string    `json:"uuid"`
		Key   string       `json:"key"`
		Created  time.Time `json:"created"`
		Modified time.Time `json:"modified"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		uuid := chi.URLParam(r, "uuid")

		cak, err := a.Service.GenerateCustomerDevKey(r.Context(), uuid)
		if err != nil {
			fmt.Fprintf(os.Stderr, "service GenerateCustomerAPIKey(ctx, %q) error: %v", uuid, err)
			return
		}

		akresp := apiDevResponse{
			Key:      cak.Key,
			Created:  cak.Created,
			Modified: cak.Modified,
		}
		w.WriteHeader(http.StatusCreated) // 201 Created
		json.NewEncoder(w).Encode(akresp)
	}
}
