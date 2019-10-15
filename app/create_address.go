package app

import (
	"context"
	"encoding/json"
	"net/http"

	service "bitbucket.org/andyfusniakteam/ecom-api-go/service/firebase"
	log "github.com/sirupsen/logrus"
)

type createAddressRequestBody struct {
	UserID      *string `json:"user_id"`
	Typ         *string `json:"type"`
	ContactName *string `json:"contact_name"`
	Addr1       *string `json:"addr1"`
	Addr2       *string `json:"addr2"`
	City        *string `json:"city"`
	County      *string `json:"county"`
	Postcode    *string `json:"postcode"`
	CountryCode *string `json:"country_code"`
}

func validateCreateAddressRequestMemoize() func(ctx context.Context, request *createAddressRequestBody) (bool, string) {
	// TODO: compile regular expression here
	// TODO: country_code list of acceptable values (ISO ....)
	return func(ctx context.Context, request *createAddressRequestBody) (bool, string) {
		// user_id attribute
		userID := request.UserID
		if userID == nil {
			return false, "user_id attribute must be set"
		}
		if !IsValidUUID(*userID) {
			return false, "user_id attribute must be a valid v4 UUID"
		}

		// type attribute
		typ := request.Typ
		if typ == nil {
			return false, "type attribute must be set"
		}
		if *typ != "billing" && *typ != "shipping" {
			return false, "type attribute must be set to either billing or shipping"
		}

		// contact_name attribute
		contactName := request.ContactName
		if contactName == nil {
			return false, "contact_name attribute must be set"
		}

		// addr1 attribute
		addr1 := request.Addr1
		if addr1 == nil {
			return false, "attr1 attribute must be set"
		}

		// city attribute
		city := request.City
		if city == nil {
			return false, "city attribute must be set"
		}

		// county attribute (optional)

		// postcode attribute
		postcode := request.Postcode
		if postcode == nil {
			return false, "postcode attribute must be set"
		}

		// country_code attribute
		countryCode := request.CountryCode
		if countryCode == nil {
			return false, "country_code attribute must be set"
		}

		return true, ""
	}
}

// CreateAddressHandler creates an HTTP handler that creates a new user address record.
func (a *App) CreateAddressHandler() http.HandlerFunc {
	validate := validateCreateAddressRequestMemoize()

	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		contextLogger := log.WithContext(ctx)
		contextLogger.Info("app: CreateAddressHandler started")

		request := createAddressRequestBody{}
		dec := json.NewDecoder(r.Body)
		dec.DisallowUnknownFields()
		err := dec.Decode(&request)
		if err != nil {
			clientError(w, http.StatusBadRequest, ErrCodeBadRequest, err.Error())
			return
		}
		defer r.Body.Close()

		// validate the request body
		valid, message := validate(ctx, &request)
		if !valid {
			clientError(w, http.StatusBadRequest, ErrCodeBadRequest, message)
			return
		}

		address, err := a.Service.CreateAddress(ctx, *request.UserID, *request.Typ, *request.ContactName, *request.Addr1, request.Addr2, *request.City, request.County, *request.Postcode, *request.CountryCode)
		if err == service.ErrUserNotFound {
			clientError(w, http.StatusNotFound, ErrCodeUserNotFound, "user not found")
			return
		}
		if err != nil {
			contextLogger.Errorf("app: a.Service.CreateAddress(ctx, userID=%q, typ=%q, contactName=%q, addr1=%q, addr2=%v, city=%q, county=%v, postcode=%q, countryCode=%q) failed with error: %v", *request.UserID, *request.Typ, *request.ContactName, *request.Addr1, request.Addr2, *request.City, request.County, *request.Postcode, *request.CountryCode, err)
			serverError(w, http.StatusInternalServerError, ErrCodeInternalServerError, "internal server error")
			return
		}

		contextLogger.Infof("app: address created id=%q, userID=%q, typ=%q, contactName=%q, addr1=%q, addr2=%v, city=%q, county=%v, postcode=%q, countryCode=%q", address.ID, address.UserID, address.Typ, address.ContactName, address.Addr1, address.Addr2, address.City, address.County, address.Postcode, address.CountryCode)
		w.WriteHeader(http.StatusCreated) // 201 Created
		json.NewEncoder(w).Encode(&address)
	}
}
