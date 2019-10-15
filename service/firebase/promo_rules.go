package firebase

import (
	"context"
	"time"

	"bitbucket.org/andyfusniakteam/ecom-api-go/model/postgres"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

// ErrPromoRuleNotFound error
var ErrPromoRuleNotFound = errors.New("service: promo rule not found")

// ErrPromoRuleExists error
var ErrPromoRuleExists = errors.New("service: promo rule exists")

// PromoRuleProduct for a set a product inside a promo rule.
type PromoRuleProduct struct {
	ProductID string `json:"product_id"`
}

// PromoRule for promotion rules
type PromoRule struct {
	Object             string     `json:"object"`
	ID                 string     `json:"id"`
	PromoRuleCode      string     `json:"promo_rule_code"`
	ProductID          string     `json:"product_id,omitempty"`
	ProductPath        string     `json:"product_path,omitempty"`
	ProductSKU         string     `json:"product_sku,omitempty"`
	CategoryID         string     `json:"category_id,omitempty"`
	CategoryPath       string     `json:"category_path,omitempty"`
	ShippingTariffID   string     `json:"shipping_tariff_id,omitempty"`
	ShippingTariffCode string     `json:"shipping_tariff_code,omitempty"`
	ProductSetID       string     `json:"product_set_id,omitempty"`
	Name               string     `json:"name"`
	StartAt            *time.Time `json:"start_at"`
	EndAt              *time.Time `json:"end_at"`
	Amount             int        `json:"amount"`
	TotalThreshold     *int       `json:"total_threshold"`
	Type               string     `json:"type"`
	Target             string     `json:"target"`
	Created            time.Time  `json:"created"`
	Modified           time.Time  `json:"modified"`
}

// PromoRuleProductRequestBody struct
type PromoRuleProductRequestBody struct {
	ProductID string `json:"product_id"`
}

// PromoRuleProductSetRequestBody struct
type PromoRuleProductSetRequestBody struct {
	Object string                         `json:"object"`
	Data   []*PromoRuleProductRequestBody `json:"data"`
}

// PromoRuleCreateRequestBody request body for creating a new promo rule.
type PromoRuleCreateRequestBody struct {
	PromoRuleCode    string                          `json:"promo_rule_code"`
	Name             string                          `json:"name"`
	StartAt          *time.Time                      `json:"start_at"`
	EndAt            *time.Time                      `json:"end_at"`
	Amount           *int                            `json:"amount"`
	TotalThreshold   *int                            `json:"total_threshold"`
	ProductID        *string                         `json:"product_id"`
	CategoryID       *string                         `json:"category_id"`
	ShippingTariffID *string                         `json:"shipping_tariff_id"`
	ProductSet       *PromoRuleProductSetRequestBody `json:"product_set"`
	Type             string                          `json:"type"`
	Target           string                          `json:"target"`
}

// CreatePromoRule creates a new promotion rule.
func (s *Service) CreatePromoRule(ctx context.Context, pr *PromoRuleCreateRequestBody) (*PromoRule, error) {
	contextLogger := log.WithContext(ctx)
	contextLogger.Infof("service: CreatePromoRule(ctx, ...) started")

	var row *postgres.PromoRuleJoinProductRow
	var rule PromoRule
	if pr.Target == "product" {
		contextLogger.Infof("service: promo rule target is a product")

		var err error
		row, err = s.model.CreatePromoRuleTargetProduct(ctx, *pr.ProductID, pr.PromoRuleCode, pr.Name, pr.StartAt, pr.EndAt, *pr.Amount, pr.Type, pr.Target)
		if err == postgres.ErrPromoRuleExists {
			return nil, ErrPromoRuleExists
		}
		if err == postgres.ErrProductNotFound {
			return nil, ErrProductNotFound
		}
		if err != nil {
			return nil, errors.Wrapf(err, "service: s.model.CreatePromoRuleTargetProduct(ctx, productUUID=%q, name=%q, startAt=%v, endAt=%v, amount=%d, typ=%q, target=%q)", *pr.ProductID, pr.Name, pr.StartAt, pr.EndAt, pr.Amount, pr.Type, pr.Target)
		}

		rule = PromoRule{
			Object:         "promo_rule",
			ID:             row.UUID,
			ProductID:      *row.ProductUUID,
			ProductPath:    *row.ProductPath,
			ProductSKU:     *row.ProductSKU,
			Name:           row.Name,
			StartAt:        row.StartAt,
			EndAt:          row.EndAt,
			Amount:         row.Amount,
			TotalThreshold: row.TotalThreshold,
			Type:           row.Type,
			Target:         row.Target,
			Created:        row.Created,
			Modified:       row.Modified,
		}
	} else if pr.Target == "productset" {
		contextLogger.Infof("service: promo rule target is a productset")

		productSet := make([]*postgres.PromoRuleCreateProduct, 0, len(pr.ProductSet.Data))
		for _, p := range pr.ProductSet.Data {
			product := postgres.PromoRuleCreateProduct{
				ProductUUID: p.ProductID,
			}
			productSet = append(productSet, &product)
		}

		var err error
		row, err = s.model.CreatePromoRuleTargetProductSet(ctx, productSet, pr.PromoRuleCode, pr.Name, pr.StartAt, pr.EndAt, *pr.Amount, pr.Type, pr.Target)
		if err == postgres.ErrPromoRuleExists {
			return nil, ErrPromoRuleExists
		}
		if err == postgres.ErrProductNotFound {
			return nil, ErrProductNotFound
		}
		if err != nil {
			return nil, errors.Wrapf(err, "service: s.model.CreatePromoRuleTargetProductSet(ctx, products=%v, name=%q, startAt=%v, endAt=%v, amount=%d, typ=%q, target=%q)", pr.ProductSet.Data, pr.Name, pr.StartAt, pr.EndAt, pr.Amount, pr.Type, pr.Target)
		}

		rule = PromoRule{
			Object:         "promo_rule",
			ID:             row.UUID,
			ProductSetID:   *row.ProductSetUUID,
			Name:           row.Name,
			StartAt:        row.StartAt,
			EndAt:          row.EndAt,
			Amount:         row.Amount,
			TotalThreshold: row.TotalThreshold,
			Type:           row.Type,
			Target:         row.Target,
			Created:        row.Created,
			Modified:       row.Modified,
		}
	} else if pr.Target == "category" {
		contextLogger.Infof("service: promo rule target is a category")

		var err error
		row, err = s.model.CreatePromoRuleTargetCategory(ctx, *pr.CategoryID, pr.PromoRuleCode, pr.Name, pr.StartAt, pr.EndAt, *pr.Amount, pr.Type, pr.Target)
		if err == postgres.ErrPromoRuleExists {
			return nil, ErrPromoRuleExists
		}
		if err == postgres.ErrCategoryNotFound {
			return nil, ErrCategoryNotFound
		}
		if err != nil {
			return nil, errors.Wrapf(err, "service: s.model.CreatePromoRuleTargetCategory(ctx, categoryUUID=%q, name=%q, startAt=%v, endAt=%v, amount=%d, typ=%q, target=%q)", *pr.CategoryID, pr.Name, pr.StartAt, pr.EndAt, pr.Amount, pr.Type, pr.Target)
		}

		rule = PromoRule{
			Object:         "promo_rule",
			ID:             row.UUID,
			CategoryID:     *row.CategoryUUID,
			CategoryPath:   *row.CategoryPath,
			Name:           row.Name,
			StartAt:        row.StartAt,
			EndAt:          row.EndAt,
			Amount:         row.Amount,
			TotalThreshold: row.TotalThreshold,
			Type:           row.Type,
			Target:         row.Target,
			Created:        row.Created,
			Modified:       row.Modified,
		}
	} else if pr.Target == "shipping_tariff" {
		contextLogger.Infof("service: promo rule target is a shipping_tariff")

		var err error
		row, err = s.model.CreatePromoRuleTargetShippingTariff(ctx, *pr.ShippingTariffID, pr.PromoRuleCode, pr.Name, pr.StartAt, pr.EndAt, *pr.Amount, pr.Type, pr.Target)
		if err == postgres.ErrPromoRuleExists {
			return nil, ErrPromoRuleExists
		}
		if err == postgres.ErrShippingTariffNotFound {
			return nil, ErrShippingTariffNotFound
		}
		if err != nil {
			return nil, errors.Wrapf(err, "service: s.model.CreatePromoRuleTargetShippingTariff(ctx, shippingTariffID=%q, name=%q, startAt=%v, endAt=%v, amount=%d, type=%q, target=%q)", *pr.ShippingTariffID, pr.Name, pr.StartAt, pr.EndAt, pr.Amount, pr.Type, pr.Target)
		}

		rule = PromoRule{
			Object:             "promo_rule",
			ID:                 row.UUID,
			ShippingTariffID:   *row.ShippingTariffUUID,
			ShippingTariffCode: *row.ShippingTariffCode,
			Name:               row.Name,
			StartAt:            row.StartAt,
			EndAt:              row.EndAt,
			Amount:             row.Amount,
			TotalThreshold:     row.TotalThreshold,
			Type:               row.Type,
			Target:             row.Target,
			Created:            row.Created,
			Modified:           row.Modified,
		}
	} else if pr.Target == "total" {
		contextLogger.Infof("service: promo rule target is a total")

		var err error
		row, err = s.model.CreatePromoRuleTargetTotal(ctx, *pr.TotalThreshold, pr.PromoRuleCode, pr.Name, pr.StartAt, pr.EndAt, *pr.Amount, pr.Type, pr.Target)
		if err == postgres.ErrPromoRuleExists {
			return nil, ErrPromoRuleExists
		}
		if err == postgres.ErrShippingTariffNotFound {
			return nil, ErrShippingTariffNotFound
		}
		if err != nil {
			return nil, errors.Wrapf(err, "service: s.model.CreatePromoRuleTargetShippingTariff(ctx, shippingTariffID=%q, name=%q, startAt=%v, endAt=%v, amount=%d, type=%q, target=%q)", *pr.ShippingTariffID, pr.Name, pr.StartAt, pr.EndAt, pr.Amount, pr.Type, pr.Target)
		}

		rule = PromoRule{
			Object:         "promo_rule",
			ID:             row.UUID,
			TotalThreshold: row.TotalThreshold,
			Name:           row.Name,
			StartAt:        row.StartAt,
			EndAt:          row.EndAt,
			Amount:         row.Amount,
			Type:           row.Type,
			Target:         row.Target,
			Created:        row.Created,
			Modified:       row.Modified,
		}
	}

	contextLogger.Infof("service: CreatePromoRule(ctx, ...) finished returning new rule=%v", rule)
	return &rule, nil
}

// GetPromoRule returns a single promotion rule by id.
func (s *Service) GetPromoRule(ctx context.Context, promoRuleID string) (*PromoRule, error) {
	row, err := s.model.GetPromoRule(ctx, promoRuleID)
	if err == postgres.ErrPromoRuleNotFound {
		return nil, ErrPromoRuleNotFound
	}
	if err != nil {
		return nil, errors.Wrapf(err, "service: s.model.GetPromoRule(ctx, promoRuleID=%q)", promoRuleID)
	}
	promoRule := PromoRule{
		Object:         "promo_rule",
		ID:             row.UUID,
		Name:           row.Name,
		StartAt:        row.StartAt,
		EndAt:          row.EndAt,
		Amount:         row.Amount,
		TotalThreshold: row.TotalThreshold,
		Type:           row.Type,
		Target:         row.Target,
		Created:        row.Created,
		Modified:       row.Modified,
	}
	return &promoRule, nil
}

// GetPromoRules returns a list of promo rules.
func (s *Service) GetPromoRules(ctx context.Context) ([]*PromoRule, error) {
	rows, err := s.model.GetPromoRules(ctx)
	if err == postgres.ErrPromoRuleNotFound {
		return nil, ErrPromoRuleNotFound
	}
	if err != nil {
		return nil, errors.Wrapf(err, "service: s.model.GetPromoRules(ctx) failed")
	}

	rules := make([]*PromoRule, 0, len(rows))
	for _, t := range rows {
		rule := PromoRule{
			Object:         "promo_rule",
			ID:             t.UUID,
			Name:           t.Name,
			StartAt:        t.StartAt,
			EndAt:          t.EndAt,
			Amount:         t.Amount,
			TotalThreshold: t.TotalThreshold,
			Type:           t.Type,
			Target:         t.Target,
			Created:        t.Created,
			Modified:       t.Modified,
		}
		rules = append(rules, &rule)
	}
	return rules, nil
}

// DeletePromoRule deletes a promotion rule.
func (s *Service) DeletePromoRule(ctx context.Context, promoRuleID string) error {
	err := s.model.DeletePromoRule(ctx, promoRuleID)
	if err == postgres.ErrPromoRuleNotFound {
		return ErrPromoRuleNotFound
	}
	if err != nil {
		return errors.Wrapf(err, "service: s.model.DeletePromoRule(ctx, promoRuleID=%q) failed", promoRuleID)
	}
	return nil
}
