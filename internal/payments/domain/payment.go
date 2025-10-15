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
	ID                    uuid.UUID
	OrderID               uuid.UUID
	StripePaymentIntentID string
	Status                generated.PaymentStatus
	TotalCents            int32
	Currency              string
	CreatedAt             time.Time
	PaidAt                time.Time
}
