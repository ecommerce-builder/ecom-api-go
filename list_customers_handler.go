package app

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"

	"github.com/go-chi/chi"
	log "github.com/sirupsen/logrus"
)

// ListCustomersHandler retrieves a list of customers with pagination
func (a *App) ListCustomersHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		startsAfter := chi.URLParam(r, "starts_after")
		// sa, err := strconv.Atoi(startsAfter)
		// if err != nil {
		// 	log.Errorf("failed to convert starts_after string %s to integer: %v", startsAfter, err)
		// 	w.Header().Set("Content-Type", "application/json")
		// 	w.WriteHeader(http.StatusUnprocessableEntity)
		// 	return
		// }

		size := chi.URLParam(r, "size")
		s, err := strconv.Atoi(size)
		if err != nil {
			log.Errorf("failed to convert size string %s to integer: %v", size, err)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnprocessableEntity)
			return
		}

		customers, err := a.Service.GetCustomers(ctx, s, startsAfter)
		if err != nil {
			fmt.Fprintf(os.Stderr, "service GetCustomers(ctx, %d, %d) error: %v", size, startsAfter, err)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK) // 200 OK
		json.NewEncoder(w).Encode(customers)
	}
}
