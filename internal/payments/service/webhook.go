package service

import (
	"context"
	"log"

	"github.com/stripe/stripe-go/v83"
	"github.com/talmage89/art-backend/internal/payments/repo"
)

type WebhookService struct {
	repo repo.Repo
}

func NewWebhookService(repo repo.Repo) *WebhookService {
	return &WebhookService{repo: repo}
}

func (s *WebhookService) HandleCheckoutComplete(ctx context.Context, session *stripe.CheckoutSession) {
	log.Print("session.Metadata", session.Metadata)
}
