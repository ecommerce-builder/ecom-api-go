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

// PromoRuleProduct for a set a product inside a promo rule.
type PromoRuleProduct struct {
	ProductID string `json:"product_id"`
}

// PromoRule for promotion rules
type PromoRule struct {
	Object             string     `json:"object"`
	ID                 string     `json:"id"`
	ProductID          string     `json:"product_id,omitempty"`
	ProductPath        string     `json:"product_path,omitempty"`
	ProductSKU         string     `json:"product_sku,omitempty"`
	CategoryID         string     `json:"category_id,omitempty"`
	CategoryPath       string     `json:"category_path,omitempty"`
	ShippingTarrifID   string     `json:"shipping_tarrif_id,omitempty"`
	ShippingTarrifCode string     `json:"shipping_tarrif_code,omitempty"`
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
	Name             string                          `json:"name"`
	StartAt          *time.Time                      `json:"start_at"`
	EndAt            *time.Time                      `json:"end_at"`
	Amount           *int                            `json:"amount"`
	TotalThreshold   *int                            `json:"total_threshold"`
	ProductID        *string                         `json:"product_id"`
	CategoryID       *string                         `json:"category_id"`
	ShippingTarrifID *string                         `json:"shipping_tarrif_id"`
	ProductSet       *PromoRuleProductSetRequestBody `json:"product_set"`
	Type             string                          `json:"type"`
	Target           string                          `json:"target"`
}

// CreatePromoRule creates a new promotion rule.
func (s *Service) CreatePromoRule(ctx context.Context, pr *PromoRuleCreateRequestBody) (*PromoRule, error) {
	contextLogger := log.WithContext(ctx)
	contextLogger.Infof("service: CreatePromoRule(ctx, ...) started")
	// contextLogger.WithFields(log.Fields{
	// 	"pr":             pr,
	// 	"Name":           pr.Name,
	// 	"StartAt":        pr.StartAt,
	// 	"EndAt":          pr.EndAt,
	// 	"Amount":         pr.Amount,
	// 	"TotalThreshold": pr.TotalThreshold,
	// 	"ProductID":      pr.ProductID,
	// 	"CategoryID":     pr.CategoryID,
	// 	"Type":           pr.Type,
	// 	"Target":         pr.Target,
	// }).Debug("request body")

	var row *postgres.PromoRuleJoinProductRow
	var rule PromoRule
	if pr.Target == "product" {
		var err error
		row, err = s.model.CreatePromoRuleTargetProduct(ctx, *pr.ProductID, pr.Name, pr.StartAt, pr.EndAt, *pr.Amount, pr.Type, pr.Target)
		if err != nil {
			if err == postgres.ErrProductNotFound {
				return nil, ErrProductNotFound
			}
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
		productSet := make([]*postgres.PromoRuleCreateProduct, 0, len(pr.ProductSet.Data))
		for _, p := range pr.ProductSet.Data {
			product := postgres.PromoRuleCreateProduct{
				ProductUUID: p.ProductID,
			}
			productSet = append(productSet, &product)
		}

		var err error
		row, err = s.model.CreatePromoRuleTargetProductSet(ctx, productSet, pr.Name, pr.StartAt, pr.EndAt, *pr.Amount, pr.Type, pr.Target)
		if err != nil {
			if err == postgres.ErrProductNotFound {
				return nil, ErrProductNotFound
			}
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
		var err error
		row, err = s.model.CreatePromoRuleTargetCategory(ctx, *pr.CategoryID, pr.Name, pr.StartAt, pr.EndAt, *pr.Amount, pr.Type, pr.Target)
		if err != nil {
			if err == postgres.ErrCategoryNotFound {
				return nil, ErrCategoryNotFound
			}
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
	} else if pr.Target == "shipping_tarrif" {
		var err error
		row, err = s.model.CreatePromoRuleTargetShippingTarrif(ctx, *pr.ShippingTarrifID, pr.Name, pr.StartAt, pr.EndAt, *pr.Amount, pr.Type, pr.Target)
		if err != nil {
			if err == postgres.ErrShippingTarrifNotFound {
				return nil, ErrShippingTarrifNotFound
			}
			return nil, errors.Wrapf(err, "service: s.model.CreatePromoRuleTargetShippingTarrif(ctx, shippingTarrifID=%q, name=%q, startAt=%v, endAt=%v, amount=%d, type=%q, target=%q)", *pr.ShippingTarrifID, pr.Name, pr.StartAt, pr.EndAt, pr.Amount, pr.Type, pr.Target)
		}

		rule = PromoRule{
			Object:             "promo_rule",
			ID:                 row.UUID,
			ShippingTarrifID:   *row.ShippingTarrifUUID,
			ShippingTarrifCode: *row.ShippingTarrifCode,
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
		var err error
		row, err = s.model.CreatePromoRuleTargetTotal(ctx, *pr.TotalThreshold, pr.Name, pr.StartAt, pr.EndAt, *pr.Amount, pr.Type, pr.Target)
		if err != nil {
			if err == postgres.ErrShippingTarrifNotFound {
				return nil, ErrShippingTarrifNotFound
			}
			return nil, errors.Wrapf(err, "service: s.model.CreatePromoRuleTargetShippingTarrif(ctx, shippingTarrifID=%q, name=%q, startAt=%v, endAt=%v, amount=%d, type=%q, target=%q)", *pr.ShippingTarrifID, pr.Name, pr.StartAt, pr.EndAt, pr.Amount, pr.Type, pr.Target)
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

	return &rule, nil
}

// GetPromoRule returns a single promotion rule by id.
func (s *Service) GetPromoRule(ctx context.Context, promoRuleID string) (*PromoRule, error) {
	row, err := s.model.GetPromoRule(ctx, promoRuleID)
	if err != nil {
		if err == postgres.ErrPromoRuleNotFound {
			return nil, ErrPromoRuleNotFound
		}
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
	if err != nil {
		if err == postgres.ErrPromoRuleNotFound {
			return nil, ErrPromoRuleNotFound
		}
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
	if err != nil {
		if err == postgres.ErrPromoRuleNotFound {
			return ErrPromoRuleNotFound
		}
		return errors.Wrapf(err, "service: s.model.DeletePromoRule(ctx, promoRuleID=%q) failed", promoRuleID)
	}
	return nil
}
