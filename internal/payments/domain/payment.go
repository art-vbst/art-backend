package domain

import (
	"time"

	"github.com/google/uuid"
	"github.com/talmage89/art-backend/internal/platform/db/generated"
)

type Payment struct {
	ID                    uuid.UUID
	StripePaymentIntentID string
	Status                generated.PaymentStatus
	TotalCents            int32
	Currency              string
	CreatedAt             time.Time
	PaidAt                time.Time
}
