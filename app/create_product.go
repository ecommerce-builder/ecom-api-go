package app

import (
	"encoding/json"
	"net/http"

	service "bitbucket.org/andyfusniakteam/ecom-api-go/service/firebase"
	log "github.com/sirupsen/logrus"
)

func validateProductCreateRequestBody(request *service.ProductCreateRequestBody) (bool, string) {
	if request.Path == "" {
		return false, "path attribute not set"
	}

	if request.SKU == "" {
		return false, "sku attribute not set"
	}

	if request.Name == "" {
		return false, "name attribute not set"
	}

	return true, ""
}

// CreateProductHandler creates a new product
func (a *App) CreateProductHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		contextLogger := log.WithContext(ctx)
		contextLogger.Info("app: CreateProductHandler called")

		request := service.ProductCreateRequestBody{}
		dec := json.NewDecoder(r.Body)
		dec.DisallowUnknownFields()
		if err := dec.Decode(&request); err != nil {
			clientError(w, http.StatusBadRequest, ErrCodeBadRequest, err.Error())
			return
		}
		defer r.Body.Close()

		valid, message := validateProductCreateRequestBody(&request)
		if !valid {
			clientError(w, http.StatusBadRequest, ErrCodeBadRequest, message)
			return
		}

		userID := ctx.Value(ecomUIDKey).(string)
		contextLogger.Debugf("app: ecom_uid=%q", userID)
		product, err := a.Service.CreateProduct(ctx, userID, &request)
		if err == service.ErrPriceListNotFound {
			clientError(w, http.StatusConflict, ErrCodePriceListNotFound, "price list could not be found")
			return
		}
		if err == service.ErrProductPathExists {
			clientError(w, http.StatusConflict, ErrCodeProductPathExists, "product path already exists")
			return
		}
		if err == service.ErrProductSKUExists {
			clientError(w, http.StatusConflict, ErrCodeProductSKUExists, "product sku already exists")
			return
		}
		if err != nil {
			contextLogger.Errorf("create product failed: %+v", err)
			w.WriteHeader(http.StatusInternalServerError) // 500 Internal Server Error
			return
		}
		contextLogger.Infof("app: product %q created", product.ID)

		w.WriteHeader(http.StatusCreated) // 201 Created
		json.NewEncoder(w).Encode(product)
	}
}
