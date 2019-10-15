package firebase

import (
	"context"
	"time"

	"bitbucket.org/andyfusniakteam/ecom-api-go/model/postgres"
	"github.com/pkg/errors"
)

// ErrOfferNotFound error
var ErrOfferNotFound = errors.New("service: offer not found")

// ErrOfferExists error
var ErrOfferExists = errors.New("service: offer exists")

// Offer struct
type Offer struct {
	Object      string    `json:"object"`
	ID          string    `json:"id"`
	PromoRuleID string    `json:"promo_rule_id"`
	Created     time.Time `json:"created"`
	Modified    time.Time `json:"modfied"`
}

// ActivateOffer creates an offer from a promo rule.
func (s *Service) ActivateOffer(ctx context.Context, promoRuleID string) (*Offer, error) {
	prow, err := s.model.AddOffer(ctx, promoRuleID)
	if err == postgres.ErrPromoRuleNotFound {
		return nil, ErrPromoRuleNotFound
	}
	if err == postgres.ErrOfferExists {
		return nil, ErrOfferExists
	}
	if err != nil {
		return nil, errors.Wrapf(err, "service: s.model.AddOffer(ctx, promoRuleUUID=%q) failed", promoRuleID)
	}

	err = s.model.CalcOfferPrices(ctx)
	if err == postgres.ErrCategoryNotFound {
		return nil, ErrCategoryNotFound
	}
	if err != nil {
		return nil, errors.Wrap(err, "service: s.model.CalcOfferPrices(ctx) failed")
	}

	offer := Offer{
		Object:      "offer",
		ID:          prow.UUID,
		PromoRuleID: prow.PromoRuleUUID,
		Created:     prow.Created,
		Modified:    prow.Modified,
	}
	return &offer, nil
}

// GetOffer returns an offer by offer id.
func (s *Service) GetOffer(ctx context.Context, offerID string) (*Offer, error) {
	prow, err := s.model.GetOfferByUUID(ctx, offerID)
	if err == postgres.ErrOfferNotFound {
		return nil, ErrOfferNotFound
	}
	if err != nil {
		return nil, errors.Wrapf(err, "service: s.model.GetOfferByUUID(ctx, offerUUID=%q)", offerID)
	}

	offer := Offer{
		Object:      "offer",
		ID:          prow.UUID,
		PromoRuleID: prow.PromoRuleUUID,
		Created:     prow.Created,
		Modified:    prow.Modified,
	}
	return &offer, nil
}

// GetOffers returns a slice of offers.
func (s *Service) GetOffers(ctx context.Context) ([]*Offer, error) {
	rows, err := s.model.GetOffers(ctx)
	if err != nil {
		return nil, errors.Wrapf(err, "service: s.model.GetPriceLists(ctx) failed")
	}

	offers := make([]*Offer, 0, len(rows))
	for _, row := range rows {
		offer := Offer{
			Object:      "offer",
			ID:          row.UUID,
			PromoRuleID: row.PromoRuleUUID,
			Created:     row.Created,
			Modified:    row.Modified,
		}
		offers = append(offers, &offer)
	}
	return offers, nil
}

// DeactivateOffer deactivates an existing offer.
func (s *Service) DeactivateOffer(ctx context.Context, offerID string) error {
	err := s.model.DeleteOfferByUUID(ctx, offerID)
	if err == postgres.ErrOfferNotFound {
		return ErrOfferNotFound
	}
	if err != nil {
		return errors.Wrapf(err, "s.model.DeleteOfferByUUID(ctx, offerUUID=%q)", offerID)
	}
	return nil
}
