package app

import (
	"encoding/json"
	"net/http"

	service "bitbucket.org/andyfusniakteam/ecom-api-go/service/firebase"
	log "github.com/sirupsen/logrus"
)

// ppAssocGroupRequestBody request body
type ppAssocGroupRequestBody struct {
	Code *string `json:"pp_assoc_group_code"`
	Name *string `json:"name"`
}

func validateAddProductToProductAssocGroupRequest(request *ppAssocGroupRequestBody) (bool, string) {
	if request.Code == nil {
		return false, "product_assoc_group_code attribute is required"
	}

	if len(*request.Code) < 3 || len(*request.Code) > 16 {
		return false, "product_assoc_group_code attribute must be between 3 and 16 characters in length"
	}

	if request.Name == nil {
		return false, "name attribute is required"
	}

	return true, ""
}

// CreatePPAssocGroupHandler returns a handler that adds
// a new product to product association group.
func (a *App) CreatePPAssocGroupHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		contextLogger := log.WithContext(ctx)
		contextLogger.Info("app: CreatePPAssocGroupHandler called")

		request := ppAssocGroupRequestBody{}
		dec := json.NewDecoder(r.Body)
		dec.DisallowUnknownFields()
		if err := dec.Decode(&request); err != nil {
			clientError(w, http.StatusBadRequest, ErrCodeBadRequest,
				err.Error()) // 400
			return
		}

		valid, message := validateAddProductToProductAssocGroupRequest(&request)
		if !valid {
			clientError(w, http.StatusBadRequest, ErrCodeBadRequest,
				message)
			log.Warnf("returning 400 bad request with %q", message)
			return
		}

		pToPAssocGroup, err := a.Service.CreateProductToProductAssocGroup(ctx, *request.Code, *request.Name)
		if err == service.ErrPPAssocGroupExists {
			clientError(w, http.StatusConflict, ErrCodePPAssocGroupExists,
				"product to product assoc group code is already exists") // 409
			return
		}
		if err != nil {
			contextLogger.Errorf("app: a.Service.CreateProductToProductAssocGroup(ctx, code=%q, name=%q) failed: %+v", *request.Code, *request.Name, err)
			w.WriteHeader(http.StatusInternalServerError) // 500
			return
		}
		w.WriteHeader(http.StatusCreated) // 201 Created
		json.NewEncoder(w).Encode(pToPAssocGroup)
	}
}
