package app

import (
	"context"
	"encoding/json"
	"net/http"

	service "bitbucket.org/andyfusniakteam/ecom-api-go/service/firebase"
	"github.com/go-chi/chi"
	log "github.com/sirupsen/logrus"
)

type updateAddressRequest struct {
	Typ         *string `json:"type"`
	ContactName *string `json:"contact_name"`
	Addr1       *string `json:"addr1"`
	Addr2       *string `json:"addr2"`
	City        *string `json:"city"`
	County      *string `json:"county"`
	Postcode    *string `json:"postcode"`
	CountryCode *string `json:"country_code"`
}

// UpdateAddressHandler updates an address by addresss UUID
func (a *App) UpdateAddressHandler() http.HandlerFunc {
	validate := validateAddressRequestMemoize()

	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		contextLogger := log.WithContext(ctx)
		contextLogger.Info("app: UpdateAddressHandler started")

		addressID := chi.URLParam(r, "id")
		if addressID == "" {
			contextLogger.Warn("app: URL param id not set")
			clientError(w, http.StatusBadRequest, ErrCodeBadRequest,
				"URL parameter id must be set") // 400
			return
		} else if !IsValidUUID(addressID) {
			contextLogger.Warn("app: URL param is not a valid v4 uuid")
			clientError(w, http.StatusBadRequest, ErrCodeBadRequest,
				"URL parameter id must be set to a valid v4 uuid") // 400
			return
		}

		request := updateAddressRequest{}
		dec := json.NewDecoder(r.Body)
		dec.DisallowUnknownFields()
		err := dec.Decode(&request)
		if err != nil {
			contextLogger.Warn("app: 400 Bad Request - decode request body failed")
			clientError(w, http.StatusBadRequest, ErrCodeBadRequest,
				err.Error()) // 400
			return
		}
		defer r.Body.Close()

		// validate the request body
		valid, message := validate(ctx, &request)
		if !valid {
			contextLogger.Warnf("app: 400 Bad Request - validate failed message=%q",
				message)
			clientError(w, http.StatusBadRequest, ErrCodeBadRequest,
				message) // 400
			return
		}

		address, err := a.Service.PartialUpdateAddress(ctx, addressID, request.Typ, request.ContactName, request.Addr1, request.Addr2, request.City, request.County, request.Postcode, request.CountryCode)
		if err == service.ErrAddressNotFound {
			contextLogger.Warnf("app: 404 Not Found - address %s not found",
				addressID)
			clientError(w, http.StatusNotFound, ErrCodeAddressNotFound,
				"address not found") // 404
			return
		}
		if err != nil {
			contextLogger.Errorf("app: a.Service.PartialUpdateAddress(ctx, request.Typ=%v) failed: %+v", request.Typ, err)
			w.WriteHeader(http.StatusInternalServerError) // 500
			return
		}

		contextLogger.Infof("app: 200 OK- address updated")
		w.WriteHeader(http.StatusOK) // 200
		json.NewEncoder(w).Encode(&address)
	}
}

func validateAddressRequestMemoize() func(ctx context.Context, request *updateAddressRequest) (bool, string) {
	// TODO: compile regexp
	return func(ctx context.Context, request *updateAddressRequest) (bool, string) {
		var atLeastOne bool

		// type attribute
		typ := request.Typ
		if typ != nil {
			if *typ != "billing" && *typ != "shipping" {
				return false, "type attribute must be set to a value of billing or shipping"
			}
			atLeastOne = true
		}

		// contact_name
		contactName := request.ContactName
		if contactName != nil {
			atLeastOne = true
		}

		// addr1
		if request.Addr1 != nil {
			atLeastOne = true
		}

		// addr2
		if request.Addr2 != nil {
			atLeastOne = true
		}

		// city
		if request.City != nil {
			atLeastOne = true
		}

		// county
		if request.County != nil {
			atLeastOne = true
		}

		// country_code
		if request.CountryCode != nil {
			atLeastOne = true
		}

		if !atLeastOne {
			return false, "you must set at least on attribute type, contact_name, addr1, addr2, city, county or country_code"
		}

		return true, ""
	}
}
