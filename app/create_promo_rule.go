package app

import (
	"encoding/json"
	"net/http"

	service "bitbucket.org/andyfusniakteam/ecom-api-go/service/firebase"
	log "github.com/sirupsen/logrus"
)

// CreatePromoRuleHandler creates a new product
func (a *App) CreatePromoRuleHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		contextLogger := log.WithContext(ctx)
		contextLogger.Info("app: CreatePromoRuleHandler called")

		request := service.PromoRuleCreateRequestBody{}
		dec := json.NewDecoder(r.Body)
		dec.DisallowUnknownFields()
		if err := dec.Decode(&request); err != nil {
			clientError(w, http.StatusBadRequest, ErrCodeBadRequest, err.Error()) // 400
			return
		}
		defer r.Body.Close()

		valid, message := validatePromoRuleCreateRequestBody(&request)
		if !valid {
			clientError(w, http.StatusBadRequest, ErrCodeBadRequest, message) // 400
			return
		}

		promoRule, err := a.Service.CreatePromoRule(ctx, &request)
		if err == service.ErrPromoRuleExists {
			clientError(w, http.StatusConflict, ErrCodePromoRuleExists,
				"promo rule code already exists") // 409
			return
		}
		if err == service.ErrProductNotFound {
			clientError(w, http.StatusNotFound, ErrCodeProductNotFound,
				"product not found") // 404
			return
		}
		if err == service.ErrCategoryNotFound {
			clientError(w, http.StatusNotFound, ErrCodeCategoryNotFound,
				"category not found") // 404
			return
		}
		if err == service.ErrShippingTariffNotFound {
			clientError(w, http.StatusNotFound, ErrCodeShippingTariffNotFound,
				"shipping tariff not found") // 404
			return
		}
		if err != nil {
			contextLogger.Errorf("app: a.Service.CreatePromoRule(ctx, pr=%v) error: %+v", request, err)
			w.WriteHeader(http.StatusInternalServerError) // 500
			return
		}
		w.WriteHeader(http.StatusCreated) // 201
		json.NewEncoder(w).Encode(&promoRule)
	}
}

func validatePromoRuleCreateRequestBody(request *service.PromoRuleCreateRequestBody) (bool, string) {
	if request.PromoRuleCode == "" {
		return false, "attribute promo_rule_code is empty"
	}

	if request.Name == "" {
		return false, "attribute name is empty and must be given a promotion rule name"
	}

	if request.Type != "percentage" && request.Type != "fixed" {
		return false, "attribute type must be set to a value of percentage or fixed"
	}

	if request.Target != "product" && request.Target != "productset" &&
		request.Target != "category" && request.Target != "total" && request.Target != "shipping_tariff" {
		return false, "attribute target must be set to a value of product, productset, category, total or shipping_tariff"
	}

	// A target of "product" requires that the request body contains a product_id attribute
	if request.Target == "product" {
		if request.ProductID == nil {
			return false, "attribute target is set to product so you must pass an additional attribute product_id"
		}

		if request.ProductSet != nil {
			return false, "target is set to product so attribute product_set must not be set"
		}
		if request.CategoryID != nil {
			return false, "target is set to product so attribute category_id must not be set"
		}
		if request.TotalThreshold != nil {
			return false, "target is set to product so attribute total_threshold must not be set"
		}
		if request.ShippingTariffID != nil {
			return false, "target is set to product so attribute shipping_tariff_id must not be set"
		}

		if !IsValidUUID(*request.ProductID) {
			return false, "product_id attribute must be a valid v4 UUID"
		}
	} else if request.Target == "productset" {
		if request.ProductSet == nil {
			return false, "target is set to product_set so attribute a list of products must be set"
		}

		if request.ProductID != nil {
			return false, "target is set to product_set so attribute product_id must not be set"
		}
		if request.CategoryID != nil {
			return false, "target is set to product_set so attribute category_id must not be set"
		}
		if request.TotalThreshold != nil {
			return false, "target is set to product_set so attribute total_threshold must not be set"
		}
		if request.ShippingTariffID != nil {
			return false, "target is set to product_set so attribute shipping_tariff_id must not be set"
		}

		// TODO: build a map and validate on duplicates

		if len(request.ProductSet.Data) == 0 {
			return false, "product_set contains an empty list of products"
		}

		for _, p := range request.ProductSet.Data {
			if !IsValidUUID(p.ProductID) {
				return false, "one or more product_set.data[..].product_id contain invalid v4 UUID"
			}
		}
	} else if request.Target == "category" {
		if request.CategoryID == nil {
			return false, "attribute target is set to category so you must pass an additional attribute category_id"
		}

		if request.ProductID != nil {
			return false, "target is set to category so attribute product_id must not be set"
		}
		if request.ProductSet != nil {
			return false, "target is set to category so attribute product_set must not be set"
		}
		if request.TotalThreshold != nil {
			return false, "target is set to category so attribute total_threshold must not be set"
		}
		if request.ShippingTariffID != nil {
			return false, "target is set to category so attribute shipping_tariff_id must not be set"
		}

		if !IsValidUUID(*request.CategoryID) {
			return false, "category_id attribute must be a valid v4 UUID"
		}
	} else if request.Target == "shipping_tariff" {
		if request.ShippingTariffID == nil {
			return false, "attribute target is set to shipping_tariff so you must pass an additional attribute shipping_tariff_id"
		}

		if request.ProductID != nil {
			return false, "target is set to shipping_tariff so attribute product_id must not be set"
		}
		if request.ProductSet != nil {
			return false, "target is set to shipping_tariff so attribute product_set must not be set"
		}
		if request.CategoryID != nil {
			return false, "target is set to shipping_tariff so attribute category_id must not be set"
		}
		if request.TotalThreshold != nil {
			return false, "target is set to shipping_tariff so attribute total_threshold must not be set"
		}

		if !IsValidUUID(*request.ShippingTariffID) {
			return false, "shipping_tariff_id attribute must be a valid v4 UUID"
		}
	} else if request.Target == "total" {
		if request.TotalThreshold == nil {
			return false, "attribute target is set to total so you must pass an additional attribute total_threshold"
		}

		if request.ProductID != nil {
			return false, "target is set to total so attribute product_id must not be set"
		}
		if request.ProductSet != nil {
			return false, "target is set to total so attribute product_set must not be set"
		}
		if request.CategoryID != nil {
			return false, "target is set to total so attribute category_id must not be set"
		}
		if request.ShippingTariffID != nil {
			return false, "target is set to total so attribute shipping_tariff must not be set"
		}
	}

	// Amount (depends on the type="percentage" or type="fixed")
	if request.Amount == nil {
		return false, "attribute amount must be set"
	}
	if *request.Amount < 0 {
		return false, "attribute amount must contain a value greater than or equal to zero"
	}
	if request.Type == "percentage" {
		if *request.Amount > 10000 {
			return false, "attribute amount must be between 0 and 10000 (0.00% to 100.00%)"
		}
	}

	return true, ""
}
