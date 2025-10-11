package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"

	"github.com/google/uuid"
	"github.com/stripe/stripe-go/v83"
	"github.com/stripe/stripe-go/v83/checkout/session"
	"github.com/talmage89/art-backend/internal/artwork/domain"
	"github.com/talmage89/art-backend/internal/artwork/repo"
	"github.com/talmage89/art-backend/internal/platform/config"
)

var (
	ErrInvalidUUIDs     = errors.New("invalid artwork UUID format")
	ErrArtworksNotFound = errors.New("one or more artworks not found")
	ErrEmptyRequest     = errors.New("artwork_ids cannot be empty")
	ErrTooManyItems     = errors.New("too many items in cart")
)

type CheckoutService struct {
	repo   repo.Repo
	config *config.Config
}

func NewCheckoutService(repo repo.Repo, config *config.Config) *CheckoutService {
	return &CheckoutService{
		repo:   repo,
		config: config,
	}
}

const MaxCheckoutItems = 50

func (s *CheckoutService) CreateCheckoutSession(ctx context.Context, artworkIdStrings []string) (*string, error) {
	if err := s.validateRequest(artworkIdStrings); err != nil {
		return nil, err
	}

	artworkIds, err := s.parseUUIDs(artworkIdStrings)
	if err != nil {
		return nil, err
	}

	artworks, err := s.fetchArtworkData(ctx, artworkIds)
	if err != nil {
		return nil, err
	}

	stripeSession, err := s.createStripeSession(artworks, artworkIds)
	if err != nil {
		return nil, err
	}

	return &stripeSession.URL, nil
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

func (s *CheckoutService) parseUUIDs(stringIds []string) ([]uuid.UUID, error) {
	ids := make([]uuid.UUID, 0, len(stringIds))

	for _, idString := range stringIds {
		id, err := uuid.Parse(idString)
		if err != nil {
			return nil, fmt.Errorf("%w: %s", ErrInvalidUUIDs, idString)
		}

		ids = append(ids, id)
	}

	return ids, nil
}

func (s *CheckoutService) fetchArtworkData(ctx context.Context, artworkIds []uuid.UUID) ([]domain.Artwork, error) {
	rows, err := s.repo.GetArtworkCheckoutData(ctx, artworkIds)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch artworks: %w", err)
	}

	if len(rows) != len(artworkIds) {
		return nil, ErrArtworksNotFound
	}

	return rows, nil
}

func (s *CheckoutService) createStripeSession(artworks []domain.Artwork, artworkIds []uuid.UUID) (*stripe.CheckoutSession, error) {
	stripe.Key = s.config.StripeSecretKey

	lineItems := s.buildLineItems(artworks)

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

func (s *CheckoutService) buildLineItems(artworks []domain.Artwork) []*stripe.CheckoutSessionLineItemParams {
	lineItems := make([]*stripe.CheckoutSessionLineItemParams, 0, len(artworks))

	for _, artwork := range artworks {
		imageURL := ""
		if len(artwork.Images) > 0 {
			imageURL = artwork.Images[0].ImageURL
		}

		productData := stripe.CheckoutSessionLineItemPriceDataProductDataParams{
			Name:   stripe.String(artwork.Title),
			Images: stripe.StringSlice([]string{imageURL}),
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

func (s *CheckoutService) buildMetadata(artworkIds []uuid.UUID) (map[string]string, error) {
	artworkIdsJSON, err := json.Marshal(artworkIds)
	if err != nil {
		return nil, err
	}

	return map[string]string{
		"artwork_ids": string(artworkIdsJSON),
	}, nil
}
