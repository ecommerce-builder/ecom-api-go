package app

import (
	"encoding/json"
	"net/http"
	"strconv"

	log "github.com/sirupsen/logrus"
)

// ListCustomersHandler retrieves a list of customers with pagination
func (a *App) ListCustomersHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		limit := r.URL.Query().Get("limit")
		l, err := strconv.Atoi(limit)
		if err != nil {
			log.Errorf("failed to convert limit string %q to integer: %v", limit, err)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnprocessableEntity)
			return
		}

		pagq := &PaginationQuery{
			OrderBy:    r.URL.Query().Get("order_by"),
			OrderDir:   r.URL.Query().Get("order_dir"),
			Limit:      l,
			StartAfter: r.URL.Query().Get("start_after"),
		}
		prs, err := a.Service.GetCustomers(r.Context(), pagq)
		if err != nil {
			log.Errorf("service GetCustomers(ctx) error: %v", err)
			return
		}

		w.WriteHeader(http.StatusOK) // 200 OK
		json.NewEncoder(w).Encode(*prs)
	}
}
