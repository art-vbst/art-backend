package domain

import (
	"time"

	"github.com/art-vbst/art-backend/internal/platform/db/generated"
	"github.com/google/uuid"
)

type PaymentStatus = generated.PaymentStatus

const (
	PaymentStatusSuccess  = generated.PaymentStatusSuccess
	PaymentStatusFailed   = generated.PaymentStatusFailed
	PaymentStatusRefunded = generated.PaymentStatusRefunded
)

type Payment struct {
	ID                    uuid.UUID               `json:"id"`
	OrderID               uuid.UUID               `json:"order_id"`
	StripePaymentIntentID string                  `json:"stripe_payment_intent_id"`
	Status                generated.PaymentStatus `json:"status"`
	TotalCents            int32                   `json:"total_cents"`
	Currency              string                  `json:"currency"`
	CreatedAt             time.Time               `json:"created_at"`
	PaidAt                time.Time               `json:"paid_at"`
}
