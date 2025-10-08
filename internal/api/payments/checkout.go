package payments

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stripe/stripe-go/v83"
	"github.com/stripe/stripe-go/v83/checkout/session"
	"github.com/talmage89/art-backend/internal/config"
	"github.com/talmage89/art-backend/internal/db"
)

var (
	ErrInvalidUUIDs     = errors.New("invalid artwork UUID format")
	ErrArtworksNotFound = errors.New("one or more artworks not found")
	ErrEmptyRequest     = errors.New("artwork_ids cannot be empty")
	ErrTooManyItems     = errors.New("too many items in cart")
)

const MaxCheckoutItems = 50

type CheckoutRequest struct {
	ArtworkIds []string `json:"artwork_ids"`
}

type CheckoutResult struct {
	URL string `json:"url"`
}

type CheckoutService struct {
	queries db.Querier
	config  *config.Config
}

func NewCheckoutService(queries db.Querier, config *config.Config) *CheckoutService {
	return &CheckoutService{
		queries: queries,
		config:  config,
	}
}

func (s *CheckoutService) CreateCheckoutSession(ctx context.Context, artworkIdStrings []string) (*CheckoutResult, error) {
	if err := s.validateRequest(artworkIdStrings); err != nil {
		return nil, err
	}

	artworkIds, err := s.parseUUIDs(artworkIdStrings)
	if err != nil {
		return nil, err
	}

	artworkData, err := s.fetchArtworkData(ctx, artworkIds)
	if err != nil {
		return nil, err
	}

	stripeSession, err := s.createStripeSession(artworkData, artworkIds)
	if err != nil {
		return nil, err
	}

	return &CheckoutResult{URL: stripeSession.URL}, nil
}

func (s *CheckoutService) validateRequest(artworkIds []string) error {
	if len(artworkIds) == 0 {
		return ErrEmptyRequest
	}

	if len(artworkIds) > MaxCheckoutItems {
		return ErrTooManyItems
	}

	return nil
}

func (s *CheckoutService) parseUUIDs(stringIds []string) ([]pgtype.UUID, error) {
	ids := make([]pgtype.UUID, 0, len(stringIds))

	for _, idString := range stringIds {
		parsedUUID, err := uuid.Parse(idString)
		if err != nil {
			return nil, fmt.Errorf("%w: %s", ErrInvalidUUIDs, idString)
		}

		id := pgtype.UUID{
			Bytes: parsedUUID,
			Valid: true,
		}

		ids = append(ids, id)
	}

	return ids, nil
}

func (s *CheckoutService) fetchArtworkData(ctx context.Context, artworkIds []pgtype.UUID) ([]db.GetStripeDataByArtworkIDsRow, error) {
	rows, err := s.queries.GetStripeDataByArtworkIDs(ctx, artworkIds)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch artworks: %w", err)
	}

	if len(rows) != len(artworkIds) {
		return nil, ErrArtworksNotFound
	}

	return rows, nil
}

func (s *CheckoutService) createStripeSession(artworkData []db.GetStripeDataByArtworkIDsRow, artworkIds []pgtype.UUID) (*stripe.CheckoutSession, error) {
	stripe.Key = s.config.StripeSecretKey

	lineItems := s.buildLineItems(artworkData)

	metadata, err := s.buildMetadata(artworkIds)
	if err != nil {
		return nil, fmt.Errorf("failed to build metadata: %w", err)
	}

	params := &stripe.CheckoutSessionParams{
		LineItems:  lineItems,
		Mode:       stripe.String(string(stripe.CheckoutSessionModePayment)),
		SuccessURL: stripe.String(s.config.FrontendUrl + "/checkout?success=true"),
		CancelURL:  stripe.String(s.config.FrontendUrl),
		Metadata:   metadata,
	}

	stripeSession, err := session.New(params)
	if err != nil {
		log.Printf("stripe session creation failed: %v", err)
		return nil, fmt.Errorf("failed to create stripe session: %w", err)
	}

	return stripeSession, nil
}

func (s *CheckoutService) buildLineItems(artworkData []db.GetStripeDataByArtworkIDsRow) []*stripe.CheckoutSessionLineItemParams {
	lineItems := make([]*stripe.CheckoutSessionLineItemParams, 0, len(artworkData))

	for _, artwork := range artworkData {
		productData := stripe.CheckoutSessionLineItemPriceDataProductDataParams{
			Name:   stripe.String(artwork.Title),
			Images: stripe.StringSlice([]string{artwork.ImageUrl}),
		}

		priceData := stripe.CheckoutSessionLineItemPriceDataParams{
			Currency:    stripe.String("usd"),
			UnitAmount:  stripe.Int64(int64(artwork.PriceCents)),
			ProductData: &productData,
		}

		lineItem := stripe.CheckoutSessionLineItemParams{
			PriceData: &priceData,
			Quantity:  stripe.Int64(1),
		}

		lineItems = append(lineItems, &lineItem)
	}

	return lineItems
}

func (s *CheckoutService) buildMetadata(artworkIds []pgtype.UUID) (map[string]string, error) {
	artworkIdsJSON, err := json.Marshal(artworkIds)
	if err != nil {
		return nil, err
	}

	return map[string]string{
		"artwork_ids": string(artworkIdsJSON),
	}, nil
}
