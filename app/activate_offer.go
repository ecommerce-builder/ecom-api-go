package app

import (
	"encoding/json"
	"net/http"

	service "bitbucket.org/andyfusniakteam/ecom-api-go/service/firebase"
	log "github.com/sirupsen/logrus"
)

type activateOfferRequest struct {
	PromoRuleID string `json:"promo_rule_id"`
}

func validateActivateOfferRequest(request *activateOfferRequest) (bool, string) {
	if request.PromoRuleID == "" {
		return false, "promo_rule_id attribute must be set"
	}

	if !IsValidUUID(request.PromoRuleID) {
		return false, "promo_rule_id is not a valid v4 UUID"
	}

	return true, ""
}

// ActivateOfferHandler creates a new product
func (a *App) ActivateOfferHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		contextLogger := log.WithContext(ctx)
		contextLogger.Info("app: ActivateOfferHandler called")

		request := activateOfferRequest{}
		dec := json.NewDecoder(r.Body)
		dec.DisallowUnknownFields()
		if err := dec.Decode(&request); err != nil {
			clientError(w, http.StatusBadRequest, ErrCodeBadRequest, err.Error())
			return
		}

		valid, message := validateActivateOfferRequest(&request)
		if !valid {
			clientError(w, http.StatusBadRequest, ErrCodeBadRequest, message)
			return
		}

		priceList, err := a.Service.ActivateOffer(ctx, request.PromoRuleID)
		if err != nil {
			if err == service.ErrPriceListCodeExists {
				clientError(w, http.StatusConflict, ErrCodePriceListCodeExists, "price list is already in use")
				return
			}
		}
		w.WriteHeader(http.StatusCreated) // 201 Created
		json.NewEncoder(w).Encode(priceList)
	}
}
