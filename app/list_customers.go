package app

import (
	"encoding/json"
	"net/http"
	"net/url"
	"strconv"

	service "bitbucket.org/andyfusniakteam/ecom-api-go/service/firebase"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

func paginationQueryFromQueryParams(v url.Values) (*service.PaginationQuery, error) {
	var limit int
	var err error
	if v.Get("limit") != "" {
		limit, err = strconv.Atoi(v.Get("limit"))
		if err != nil {
			return nil, errors.Wrapf(err, "pagination query from query params atoi of %q failed", v.Get("limit"))
		}
	} else {
		limit = 0 // unlimited
	}
	pq := &service.PaginationQuery{
		OrderBy:    v.Get("order_by"),
		OrderDir:   v.Get("order_dir"),
		Limit:      limit,
		StartAfter: v.Get("start_after"),
	}
	return pq, nil
}

// ListCustomersHandler returns an http.HandlerFunc that call the service
// API to retrievs a list of Customers using a PaginationQuery.
func (a *App) ListCustomersHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		contextLogger := log.WithContext(ctx)
		contextLogger.Info("App: ListCustomersHandler started")

		pq, err := paginationQueryFromQueryParams(r.URL.Query())
		if err != nil {
			log.Errorf("pagination query (query params=%s) from query params failed: %v", r.URL, err)
			w.WriteHeader(http.StatusUnprocessableEntity)
			return
		}
		//pagq := &PaginationQuery{
		//	OrderBy:    r.URL.Query().Get("order_by"),
		//	OrderDir:   r.URL.Query().Get("order_dir"),
		//	Limit:      l,
		//	StartAfter: r.URL.Query().Get("start_after"),
		//}
		prs, err := a.Service.GetCustomers(ctx, pq)
		if err != nil {
			log.Errorf("service GetCustomers(ctx) error: %+v", err)
			return
		}

		w.WriteHeader(http.StatusOK) // 200 OK
		json.NewEncoder(w).Encode(*prs)
	}
}
