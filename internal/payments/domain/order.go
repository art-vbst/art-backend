package domain

import (
	"time"

	"github.com/art-vbst/art-backend/internal/platform/db/generated"
	"github.com/google/uuid"
)

const (
	DefaultShippingCents = 1_000
	DefaultCurrency      = "usd"
)

type OrderStatus = generated.OrderStatus

const (
	OrderStatusPending    OrderStatus = "pending"
	OrderStatusProcessing OrderStatus = "processing"
	OrderStatusShipped    OrderStatus = "shipped"
	OrderStatusCompleted  OrderStatus = "completed"
	OrderStatusFailed     OrderStatus = "failed"
	OrderStatusCanceled   OrderStatus = "canceled"
)

type OrderPublic struct {
	ID        uuid.UUID   `json:"id"`
	Status    OrderStatus `json:"status"`
	CreatedAt time.Time   `json:"created_at"`
}

type Order struct {
	ID                 uuid.UUID          `json:"id"`
	StripeSessionID    *string            `json:"stripe_session_id,omitempty"`
	Status             OrderStatus        `json:"status"`
	ShippingDetail     ShippingDetail     `json:"shipping_detail"`
	PaymentRequirement PaymentRequirement `json:"payment_requirement"`
	Payments           []Payment          `json:"payments"`
	CreatedAt          time.Time          `json:"created_at"`
}

type ShippingDetail struct {
	ID      uuid.UUID `json:"id"`
	OrderID uuid.UUID `json:"order_id"`
	Email   string    `json:"email"`
	Name    string    `json:"name"`
	Line1   string    `json:"line1"`
	Line2   *string   `json:"line2,omitempty"`
	City    string    `json:"city"`
	State   string    `json:"state"`
	Postal  string    `json:"postal"`
	Country string    `json:"country"`
}

type PaymentRequirement struct {
	ID            uuid.UUID `json:"id"`
	OrderID       uuid.UUID `json:"order_id"`
	SubtotalCents int32     `json:"subtotal_cents"`
	ShippingCents int32     `json:"shipping_cents"`
	TotalCents    int32     `json:"total_cents"`
	Currency      string    `json:"currency"`
}
