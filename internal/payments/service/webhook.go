package service

import (
	"context"
	"errors"
	"time"

	artdomain "github.com/art-vbst/art-backend/internal/artwork/domain"
	artrepo "github.com/art-vbst/art-backend/internal/artwork/repo"
	paydomain "github.com/art-vbst/art-backend/internal/payments/domain"
	payrepo "github.com/art-vbst/art-backend/internal/payments/repo"
	"github.com/google/uuid"
	"github.com/stripe/stripe-go/v83"
)

var (
	ErrMetadataParse = errors.New("failed to parse session metadata")
	ErrOrderNotFound = errors.New("provided order id not found")
	ErrNotPaid       = errors.New("payment intent not successful")
)

type WebhookService struct {
	payrepo payrepo.Repo
	artrepo artrepo.Repo
	emails  *EmailService
}

func NewWebhookService(payrepo payrepo.Repo, artrepo artrepo.Repo, emails *EmailService) *WebhookService {
	return &WebhookService{payrepo: payrepo, artrepo: artrepo, emails: emails}
}

func (s *WebhookService) HandleCheckoutComplete(ctx context.Context, session *stripe.CheckoutSession) error {
	orderID, err := s.getMetadataOrderID(session)
	if err != nil {
		return err
	}

	if session.PaymentStatus != stripe.CheckoutSessionPaymentStatusPaid {
		return ErrNotPaid
	}

	shipping := paydomain.ShippingDetail{
		Email:   session.CustomerDetails.Email,
		Name:    session.CollectedInformation.ShippingDetails.Name,
		Line1:   session.CollectedInformation.ShippingDetails.Address.Line1,
		Line2:   &session.CollectedInformation.ShippingDetails.Address.Line2,
		City:    session.CollectedInformation.ShippingDetails.Address.City,
		State:   session.CollectedInformation.ShippingDetails.Address.State,
		Postal:  session.CollectedInformation.ShippingDetails.Address.PostalCode,
		Country: session.CollectedInformation.ShippingDetails.Address.Country,
	}

	order := paydomain.Order{
		ID:              *orderID,
		StripeSessionID: &session.ID,
		Status:          paydomain.OrderStatusProcessing,
		ShippingDetail:  shipping,
	}

	payment := paydomain.Payment{
		OrderID:               *orderID,
		StripePaymentIntentID: session.PaymentIntent.ID,
		Status:                paydomain.PaymentStatusSuccess,
		TotalCents:            int32(session.AmountTotal),
		Currency:              string(session.Currency),
		PaidAt:                time.Now(),
	}

	if err := s.payrepo.UpdateOrderWithPayment(ctx, &order, &payment); err != nil {
		return err
	}

	if err := s.artrepo.UpdateArtworkStatuses(ctx, *orderID, artdomain.ArtworkStatusSold); err != nil {
		return err
	}

	s.emails.SendOrderRecieved(order.ID, shipping.Email)

	return nil
}

func (s *WebhookService) HandleCheckoutExpired(ctx context.Context, session *stripe.CheckoutSession) error {
	orderID, err := s.getMetadataOrderID(session)
	if err != nil {
		return err
	}

	if err := s.artrepo.UpdateArtworkStatuses(ctx, *orderID, artdomain.ArtworkStatusAvailable); err != nil {
		return err
	}

	if err := s.payrepo.DeleteOrder(ctx, *orderID); err != nil {
		return err
	}

	return nil
}

func (s *WebhookService) getMetadataOrderID(session *stripe.CheckoutSession) (*uuid.UUID, error) {
	val, ok := session.Metadata["order_id"]
	if !ok || val == "" {
		return nil, ErrMetadataParse
	}

	id, err := uuid.Parse(val)
	if err != nil {
		return nil, ErrMetadataParse
	}

	return &id, nil
}
