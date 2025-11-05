package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"

	artdomain "github.com/art-vbst/art-backend/internal/artwork/domain"
	artrepo "github.com/art-vbst/art-backend/internal/artwork/repo"
	paydomain "github.com/art-vbst/art-backend/internal/payments/domain"
	payrepo "github.com/art-vbst/art-backend/internal/payments/repo"
	"github.com/art-vbst/art-backend/internal/platform/config"
	"github.com/google/uuid"
	"github.com/stripe/stripe-go/v83"
	"github.com/stripe/stripe-go/v83/checkout/session"
)

var (
	ErrInvalidUUIDs     = errors.New("invalid artwork UUID format")
	ErrArtworksNotFound = errors.New("one or more artworks not found")
	ErrEmptyRequest     = errors.New("artwork_ids cannot be empty")
	ErrTooManyItems     = errors.New("too many items in cart")
	ErrMetadataParse    = errors.New("failed to parse session metadata")
)

type CheckoutService struct {
	artrepo artrepo.Repo
	payrepo payrepo.Repo
	config  *config.Config
}

func NewCheckoutService(artrepo artrepo.Repo, payrepo payrepo.Repo, config *config.Config) *CheckoutService {
	return &CheckoutService{
		artrepo: artrepo,
		payrepo: payrepo,
		config:  config,
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

	order, err := s.createOrder(ctx, artworks)
	if err != nil {
		return nil, err
	}

	session, err := s.createCheckoutSession(artworks, order.ID)
	if err != nil {
		return nil, err
	}

	return &session.URL, nil
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

func (s *CheckoutService) parseUUIDs(idStrings []string) ([]uuid.UUID, error) {
	ids := make([]uuid.UUID, 0, len(idStrings))

	for _, idString := range idStrings {
		id, err := uuid.Parse(idString)
		if err != nil {
			return nil, fmt.Errorf("%w: %s", ErrInvalidUUIDs, idString)
		}

		ids = append(ids, id)
	}

	return ids, nil
}

func (s *CheckoutService) fetchArtworkData(ctx context.Context, artworkIds []uuid.UUID) ([]artdomain.Artwork, error) {
	artworks, err := s.artrepo.GetArtworkCheckoutData(ctx, artworkIds)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch artworks: %w", err)
	}

	if len(artworks) != len(artworkIds) {
		return nil, ErrArtworksNotFound
	}

	return artworks, nil
}

func (s *CheckoutService) createOrder(ctx context.Context, artworks []artdomain.Artwork) (*paydomain.Order, error) {
	orderParams := paydomain.Order{
		Status:             paydomain.OrderStatusPending,
		PaymentRequirement: s.getOrderPaymentRequirement(artworks),
	}

	order, err := s.payrepo.CreateOrder(ctx, &orderParams)
	if err != nil {
		return nil, err
	}

	return order, nil
}

func (s *CheckoutService) getOrderPaymentRequirement(artworks []artdomain.Artwork) paydomain.PaymentRequirement {
	var subtotal int32
	for _, artwork := range artworks {
		subtotal += artwork.PriceCents
	}

	return paydomain.PaymentRequirement{
		SubtotalCents: subtotal,
		ShippingCents: paydomain.DefaultShippingCents,
		TotalCents:    subtotal + paydomain.DefaultShippingCents,
		Currency:      paydomain.DefaultCurrency,
	}
}

func (s *CheckoutService) createCheckoutSession(artworks []artdomain.Artwork, orderId uuid.UUID) (*stripe.CheckoutSession, error) {
	var (
		lineItems         = s.buildLineItems(artworks)
		shippingAddress   = s.buildShippingAddressParams()
		shippingOptions   = s.buildShippingOptionParams()
		paymentIntentData = s.buildPaymentIntentDataParams()
	)

	metadata, err := buildCheckoutSessionMetadata(artworks, orderId)
	if err != nil {
		return nil, err
	}

	params := &stripe.CheckoutSessionParams{
		Metadata:                  metadata,
		LineItems:                 lineItems,
		ShippingAddressCollection: shippingAddress,
		ShippingOptions:           shippingOptions,
		PaymentIntentData:         paymentIntentData,
		Mode:                      stripe.String(string(stripe.CheckoutSessionModePayment)),
		SuccessURL:                stripe.String(s.config.FrontendUrl + "/checkout/return?success=true&order_id=" + orderId.String()),
		CancelURL:                 stripe.String(s.config.FrontendUrl),
	}

	stripe.Key = s.config.StripeSecret
	session, err := session.New(params)
	if err != nil {
		log.Printf("stripe session creation failed: %v", err)
		return nil, fmt.Errorf("failed to create stripe session: %w", err)
	}

	return session, nil
}

func (s *CheckoutService) buildLineItems(artworks []artdomain.Artwork) []*stripe.CheckoutSessionLineItemParams {
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

func (s *CheckoutService) buildShippingAddressParams() *stripe.CheckoutSessionShippingAddressCollectionParams {
	return &stripe.CheckoutSessionShippingAddressCollectionParams{
		AllowedCountries: []*string{stripe.String("US")},
	}
}

func (s *CheckoutService) buildShippingOptionParams() []*stripe.CheckoutSessionShippingOptionParams {
	shippingOption := &stripe.CheckoutSessionShippingOptionParams{
		ShippingRateData: &stripe.CheckoutSessionShippingOptionShippingRateDataParams{
			DisplayName: stripe.String("Standard"),
			Type:        stripe.String("fixed_amount"),
			FixedAmount: &stripe.CheckoutSessionShippingOptionShippingRateDataFixedAmountParams{
				Amount:   stripe.Int64(paydomain.DefaultShippingCents),
				Currency: stripe.String(paydomain.DefaultCurrency),
			},
		},
	}
	return []*stripe.CheckoutSessionShippingOptionParams{shippingOption}
}

func (s *CheckoutService) buildPaymentIntentDataParams() *stripe.CheckoutSessionPaymentIntentDataParams {
	return &stripe.CheckoutSessionPaymentIntentDataParams{
		CaptureMethod: stripe.String(stripe.CheckoutSessionPaymentMethodOptionsAffirmCaptureMethodManual),
	}
}

type CheckoutSessionMetadata struct {
	OrderID    uuid.UUID   `json:"order_id"`
	ArtworkIDs []uuid.UUID `json:"artwork_ids"`
}

func buildCheckoutSessionMetadata(artworks []artdomain.Artwork, orderId uuid.UUID) (map[string]string, error) {
	artworkIDs := make([]uuid.UUID, len(artworks))
	for i, artwork := range artworks {
		artworkIDs[i] = artwork.ID
	}

	artworkIDsStr, err := json.Marshal(artworkIDs)
	if err != nil {
		return nil, err
	}

	return map[string]string{"order_id": orderId.String(), "artwork_ids": string(artworkIDsStr)}, nil
}

func getCheckoutSessionMetadata(session *stripe.CheckoutSession) (*CheckoutSessionMetadata, error) {
	orderIDStr, ok := session.Metadata["order_id"]
	if !ok || orderIDStr == "" {
		return nil, ErrMetadataParse
	}

	orderID, err := uuid.Parse(orderIDStr)
	if err != nil {
		return nil, ErrMetadataParse
	}

	artworkIDsStr, ok := session.Metadata["artwork_ids"]
	if !ok || artworkIDsStr == "" {
		return nil, ErrMetadataParse
	}

	var artworkIDs []uuid.UUID
	if err := json.Unmarshal([]byte(artworkIDsStr), &artworkIDs); err != nil {
		return nil, ErrMetadataParse
	}

	metadata := &CheckoutSessionMetadata{
		OrderID:    orderID,
		ArtworkIDs: artworkIDs,
	}

	return metadata, nil
}
