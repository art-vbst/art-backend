package domain

import (
	"time"

	"github.com/google/uuid"
	"github.com/talmage89/art-backend/internal/platform/db/generated"
)

type Order struct {
	ID                 uuid.UUID
	StripeSessionID    string
	Status             generated.OrderStatus
	ShippingDetails    ShippingDetails
	PaymentRequirement PaymentRequirement
	Payments           []Payment
	CreatedAt          time.Time
}

type ShippingDetails struct {
	ID      uuid.UUID
	Email   string
	Name    string
	Line1   string
	Line2   string
	City    string
	State   string
	Postal  string
	Country string
}

type PaymentRequirement struct {
	ID            uuid.UUID
	SubtotalCents int
	ShippingCents int
	TotalCents    int
	Currency      string
}

type Payment struct {
	ID                    uuid.UUID
	StripePaymentIntentID string
	Status                generated.PaymentStatus
	TotalCents            int
	Currency              string
	CreatedAt             time.Time
	PaidAt                time.Time
}
