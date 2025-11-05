package service

import (
	"context"
	"errors"
	"time"

	artrepo "github.com/art-vbst/art-backend/internal/artwork/repo"
	paydomain "github.com/art-vbst/art-backend/internal/payments/domain"
	payrepo "github.com/art-vbst/art-backend/internal/payments/repo"
	"github.com/art-vbst/art-backend/internal/platform/config"
	"github.com/google/uuid"
	"github.com/stripe/stripe-go/v83"
	"github.com/stripe/stripe-go/v83/paymentintent"
)

var (
	ErrOrderNotFound           = errors.New("provided order id not found")
	ErrBadIntentStatus         = errors.New("unexpected payment intent status during capture/cancel")
	ErrIntentAlreadyProcessed  = errors.New("payment intent has already been captured/cancled")
	ErrIntentProcessingFailure = errors.New("failed to process payment intent")
	ErrArtworksNotAvailable    = errors.New("one or more artworks not available for purchase")
)

type WebhookService struct {
	payrepo payrepo.Repo
	artrepo artrepo.Repo
	emails  *EmailService
	config  *config.Config
}

func NewWebhookService(payrepo payrepo.Repo, artrepo artrepo.Repo, emails *EmailService, config *config.Config) *WebhookService {
	return &WebhookService{payrepo: payrepo, artrepo: artrepo, emails: emails, config: config}
}

func (s *WebhookService) HandleCheckoutComplete(ctx context.Context, session *stripe.CheckoutSession) error {
	metadata, err := getCheckoutSessionMetadata(session)
	if err != nil {
		return err
	}

	order, payment := s.constructDomainData(metadata.OrderID, session)

	err = s.artrepo.UpdateArtworksAsPurchased(ctx, metadata.ArtworkIDs, metadata.OrderID, func(selectedIDs []uuid.UUID) error {
		if len(metadata.ArtworkIDs) != len(selectedIDs) {
			return ErrArtworksNotAvailable
		}

		if err := s.capturePaymentIntent(session.PaymentIntent.ID); err != nil {
			return err
		}

		if err := s.payrepo.UpdateOrderWithPayment(ctx, order, payment); err != nil {
			return err
		}

		s.emails.SendOrderRecieved(order.ID, order.ShippingDetail.Email)
		return nil
	})

	if err != nil {
		s.cleanupSessionState(ctx, metadata.OrderID, session.PaymentIntent)
		return err
	}

	return nil
}

func (s *WebhookService) HandleCheckoutExpired(ctx context.Context, session *stripe.CheckoutSession) error {
	metadata, err := getCheckoutSessionMetadata(session)
	if err != nil {
		return err
	}

	if err := s.cleanupSessionState(ctx, metadata.OrderID, session.PaymentIntent); err != nil {
		return err
	}

	return nil
}

func (s *WebhookService) cleanupSessionState(ctx context.Context, orderID uuid.UUID, paymentIntent *stripe.PaymentIntent) error {
	if paymentIntent != nil {
		if err := s.cancelPaymentIntent(paymentIntent.ID); err != nil {
			return err
		}
	}

	if err := s.payrepo.DeleteOrder(ctx, orderID); err != nil {
		return err
	}

	return nil
}

func (s *WebhookService) capturePaymentIntent(paymentIntentID string) error {
	pi, err := s.getPaymentIntentForCapture(paymentIntentID, stripe.PaymentIntentStatusSucceeded)
	if err != nil {
		return err
	}

	pi, err = paymentintent.Capture(pi.ID, nil)
	if err != nil {
		return err
	}
	if pi.Status != stripe.PaymentIntentStatusSucceeded {
		return ErrIntentProcessingFailure
	}

	return nil
}

func (s *WebhookService) cancelPaymentIntent(paymentIntentID string) error {
	pi, err := s.getPaymentIntentForCapture(paymentIntentID, stripe.PaymentIntentStatusCanceled)
	if err != nil {
		return err
	}

	pi, err = paymentintent.Cancel(pi.ID, nil)
	if err != nil {
		return err
	}
	if pi.Status != stripe.PaymentIntentStatusCanceled {
		return ErrIntentProcessingFailure
	}

	return nil
}

func (s *WebhookService) getPaymentIntentForCapture(paymentIntentID string, targetStatus stripe.PaymentIntentStatus) (*stripe.PaymentIntent, error) {
	stripe.Key = s.config.StripeSecret

	intent, err := paymentintent.Get(paymentIntentID, nil)
	if err != nil {
		return nil, err
	}
	if intent.Status == targetStatus {
		return nil, ErrIntentAlreadyProcessed
	}
	if intent.Status != stripe.PaymentIntentStatusRequiresCapture {
		return nil, ErrBadIntentStatus
	}

	return intent, nil
}

func (s *WebhookService) constructDomainData(orderID uuid.UUID, session *stripe.CheckoutSession) (*paydomain.Order, *paydomain.Payment) {
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
		ID:              orderID,
		StripeSessionID: &session.ID,
		Status:          paydomain.OrderStatusProcessing,
		ShippingDetail:  shipping,
	}

	payment := paydomain.Payment{
		OrderID:               orderID,
		StripePaymentIntentID: session.PaymentIntent.ID,
		Status:                paydomain.PaymentStatusSuccess,
		TotalCents:            int32(session.AmountTotal),
		Currency:              string(session.Currency),
		PaidAt:                time.Now(),
	}

	return &order, &payment
}
