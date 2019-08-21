package app

import (
	"encoding/json"
	"net/http"

	service "bitbucket.org/andyfusniakteam/ecom-api-go/service/firebase"
	log "github.com/sirupsen/logrus"
)

// ListPriceListsHandler creates a handler function that returns a
// list of price lists.
func (a *App) ListPriceListsHandler() http.HandlerFunc {
	type listPriceListsResponse struct {
		Object string               `json:"object"`
		Data   []*service.PriceList `json:"data"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		contextLogger := log.WithContext(ctx)
		contextLogger.Info("App: ListPriceListsHandler started")

		priceLists, err := a.Service.GetPriceLists(ctx)
		if err != nil {
			contextLogger.Errorf("service GetPricingLists(ctx) error: %+v", err)
			w.WriteHeader(http.StatusInternalServerError) // 500 Internal Server Error
			return
		}

		list := listPriceListsResponse{
			Object: "list",
			Data:   priceLists,
		}
		w.WriteHeader(http.StatusOK) // 200 OK
		json.NewEncoder(w).Encode(&list)
	}
}
