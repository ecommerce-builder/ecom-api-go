package app

import (
	"encoding/json"
	"net/http"

	service "bitbucket.org/andyfusniakteam/ecom-api-go/service/firebase"
	log "github.com/sirupsen/logrus"
)

type addressResponseBody struct {
	Object string `json:"object"`
	*service.Address
}

// CreateAddressHandler creates an HTTP handler that creates a new user address record.
func (a *App) CreateAddressHandler() http.HandlerFunc {
	type requestBody struct {
		UserID      string  `json:"user_id"`
		Typ         string  `json:"typ"`
		ContactName string  `json:"contact_name"`
		Addr1       string  `json:"addr1"`
		Addr2       *string `json:"addr2"`
		City        string  `json:"city"`
		County      *string `json:"county"`
		Postcode    string  `json:"postcode"`
		Country     string  `json:"country"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		contextLogger := log.WithContext(ctx)
		contextLogger.Info("App: CreateAddressHandler started")

		if r.Body == nil {
			http.Error(w, "Please send a request body", 400)
			return
		}

		request := requestBody{}
		err := json.NewDecoder(r.Body).Decode(&request)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest) // 400 Bad Request
			json.NewEncoder(w).Encode(struct {
				Status  int    `json:"status"`
				Code    string `json:"code"`
				Message string `json:"message"`
			}{
				http.StatusBadRequest,
				ErrCodeBadRequest,
				err.Error(),
			})
			return
		}
		defer r.Body.Close()

		address, err := a.Service.CreateAddress(ctx, request.UserID, request.Typ, request.ContactName, request.Addr1, request.Addr2, request.City, request.County, request.Postcode, request.Country)
		if err != nil {
			contextLogger.Errorf("a.Service.CreateAddress(ctx, userID=%q, ...) failed with error: %v", request.UserID, err)
			w.WriteHeader(http.StatusInternalServerError) // 500 Internal Server Error
			return
		}

		response := addressResponseBody{
			Object:  "address",
			Address: address,
		}
		w.WriteHeader(http.StatusCreated) // 201 Created
		json.NewEncoder(w).Encode(&response)
	}
}
