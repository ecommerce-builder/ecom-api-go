package app

import (
	"encoding/json"
	"net/http"

	service "bitbucket.org/andyfusniakteam/ecom-api-go/service/firebase"
	log "github.com/sirupsen/logrus"
)

// ListOffersHandler creates a handler function that returns a
// list of active offers.
func (a *App) ListOffersHandler() http.HandlerFunc {
	type response struct {
		Object string           `json:"object"`
		Data   []*service.Offer `json:"data"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		contextLogger := log.WithContext(ctx)
		contextLogger.Info("app: ListOffersHandler started")

		offers, err := a.Service.GetOffers(ctx)
		if err != nil {
			contextLogger.Errorf("app: a.Service.GetOffers(ctx) failed: %+v", err)
			w.WriteHeader(http.StatusInternalServerError) // 500 Internal Server Error
			return
		}

		list := response{
			Object: "list",
			Data:   offers,
		}
		w.WriteHeader(http.StatusOK) // 200 OK
		json.NewEncoder(w).Encode(&list)
	}
}
