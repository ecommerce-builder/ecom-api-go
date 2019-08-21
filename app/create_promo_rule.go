package app

import (
	"encoding/json"
	"net/http"

	service "bitbucket.org/andyfusniakteam/ecom-api-go/service/firebase"
	log "github.com/sirupsen/logrus"
)

func validatePromoRuleCreateRequestBody(r *service.PromoRuleCreateRequestBody) (bool, string) {
	if r.Name == "" {
		return false, "attribute name is empty and must be given a promotion rule name"
	}

	if r.Type != "percentage" && r.Type != "fixed" {
		return false, "attribute type must be set to a value of percentage or fixed"
	}

	if r.Target != "product" && r.Target != "productset" &&
		r.Target != "category" && r.Target != "total" && r.Target != "shipping" {
		return false, "attribute target must be set to a value of product, productset, category, total or shipping"
	}
	return true, ""
}

// CreatePromoRuleHandler creates a new product
func (a *App) CreatePromoRuleHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		contextLogger := log.WithContext(ctx)
		contextLogger.Info("app: CreatePromoRuleHandler called")

		requestBody := service.PromoRuleCreateRequestBody{}
		if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
			http.Error(w, err.Error(), 400)
			return
		}
		defer r.Body.Close()

		valid, message := validatePromoRuleCreateRequestBody(&requestBody)
		if !valid {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(struct {
				Status  int    `json:"status"`
				Code    string `json:"code"`
				Message string `json:"message"`
			}{
				http.StatusBadRequest,
				ErrCodeBadRequest,
				message,
			})
			return
		}

		promoRule, err := a.Service.CreatePromoRule(ctx, &requestBody)
		if err != nil {
			contextLogger.Errorf("app: a.Service.CreatePromoRule(ctx, pr=%v)) error: %+v", requestBody, err)
			w.WriteHeader(http.StatusInternalServerError) // 500 Internal Server Error
			return
		}
		w.WriteHeader(http.StatusCreated) // 201 Created
		json.NewEncoder(w).Encode(&promoRule)
	}
}
