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
	ShippingDetail     ShippingDetail
	PaymentRequirement PaymentRequirement
	Payments           []Payment
	CreatedAt          time.Time
}

type ShippingDetail struct {
	ID      uuid.UUID
	OrderID uuid.UUID
	Email   string
	Name    string
	Line1   string
	Line2   *string
	City    string
	State   string
	Postal  string
	Country string
}

type PaymentRequirement struct {
	ID            uuid.UUID
	OrderID       uuid.UUID
	SubtotalCents int32
	ShippingCents int32
	TotalCents    int32
	Currency      string
}
