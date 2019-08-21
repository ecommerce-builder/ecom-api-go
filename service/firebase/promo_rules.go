package firebase

import (
	"context"
	"time"

	"bitbucket.org/andyfusniakteam/ecom-api-go/model/postgres"
	"github.com/pkg/errors"
)

// ErrPromoRuleNotFound error
var ErrPromoRuleNotFound = errors.New("service: promo rule not found")

// PromoRule for promotion rules
type PromoRule struct {
	Object         string     `json:"object"`
	ID             string     `json:"id"`
	Name           string     `json:"name"`
	StartsAt       *time.Time `json:"starts_at"`
	EndsAt         *time.Time `json:"ends_at"`
	Amount         int        `json:"amount"`
	TotalThreshold *int       `json:"total_threshold"`
	Type           string     `json:"type"`
	Target         string     `json:"target"`
	Created        time.Time  `json:"created"`
	Modified       time.Time  `json:"modified"`
}

// PromoRuleCreateRequestBody request body for creating a new promo rule.
type PromoRuleCreateRequestBody struct {
	Name           string     `json:"name"`
	StartsAt       *time.Time `json:"starts_at"`
	EndsAt         *time.Time `json:"ends_at"`
	Amount         int        `json:"amount"`
	TotalThreshold *int       `json:"total_threshold"`
	Type           string     `json:"type"`
	Target         string     `json:"target"`
}

// CreatePromoRule creates a new promotion rule.
func (s *Service) CreatePromoRule(ctx context.Context, pr *PromoRuleCreateRequestBody) (*PromoRule, error) {
	r := postgres.PromoRuleCreate{
		Name:           pr.Name,
		StartsAt:       pr.StartsAt,
		EndsAt:         pr.EndsAt,
		Amount:         pr.Amount,
		TotalThreshold: pr.TotalThreshold,
		Type:           pr.Type,
		Target:         pr.Target,
	}
	row, err := s.model.CreatePromoRule(ctx, &r)
	if err != nil {
		return nil, errors.Wrapf(err, "service: s.model.CreatePromoRule(ctx, rule=%v)", r)
	}

	rule := PromoRule{
		Object:         "promo_rule",
		ID:             row.UUID,
		Name:           row.Name,
		StartsAt:       row.StartsAt,
		EndsAt:         row.EndsAt,
		Amount:         row.Amount,
		TotalThreshold: row.TotalThreshold,
		Type:           row.Type,
		Target:         row.Target,
		Created:        row.Created,
		Modified:       row.Modified,
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
		StartsAt:       row.StartsAt,
		EndsAt:         row.EndsAt,
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
			StartsAt:       t.StartsAt,
			EndsAt:         t.EndsAt,
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
