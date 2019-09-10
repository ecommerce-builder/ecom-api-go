package app

import (
	"encoding/json"
	"fmt"
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

	orderBy := v.Get("order_by")
	var orderDirection string
	if orderBy[0:1] == "-" {
		orderDirection = "desc"
		orderBy = orderBy[1:]
	} else {
		orderDirection = "asc"
	}

	pq := &service.PaginationQuery{
		OrderBy:    orderBy,
		OrderDir:   orderDirection,
		Limit:      limit,
		StartAfter: v.Get("start_after"),
		EndBefore:  v.Get("end_before"),
	}
	return pq, nil
}

// ListUsersHandler returns an http.HandlerFunc that call the service
// API to retrievs a list of users using a PaginationQuery.
func (a *App) ListUsersHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		contextLogger := log.WithContext(ctx)
		contextLogger.Info("app: ListUsersHandler started")

		pq, err := paginationQueryFromQueryParams(r.URL.Query())
		if err != nil {
			log.Errorf("app: pagination query (query params=%s) from query params failed: %v", r.URL, err)
			w.WriteHeader(http.StatusUnprocessableEntity)
			return
		}

		paginationResultSet, err := a.Service.GetUsers(ctx, pq)
		if err != nil {
			log.Errorf("app: GetUsers(ctx) error: %+v", err)
			return
		}

		type links struct {
			Prev string `json:"prev"`
			Next string `json:"next"`
		}

		type result struct {
			Object string      `json:"object"`
			Data   interface{} `json:"data"`
			Links  links       `json:"links"`
		}

		results := result{
			Object: "list",
			Data:   paginationResultSet.RSet,
			Links: links{
				Next: fmt.Sprintf("/users?start_after=%s", paginationResultSet.RContext.LastID),
			},
		}

		w.WriteHeader(http.StatusOK) // 200 OK
		json.NewEncoder(w).Encode(&results)
	}
}
