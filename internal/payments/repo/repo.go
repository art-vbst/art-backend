package repo

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/talmage89/art-backend/internal/payments/domain"
	"github.com/talmage89/art-backend/internal/platform/db/generated"
	"github.com/talmage89/art-backend/internal/platform/db/store"
)

type Repo interface {
	// Orders
	CreateOrder(ctx context.Context, order domain.Order) (domain.Order, error)
	GetOrderByID(ctx context.Context, id uuid.UUID) (domain.Order, error)
	GetOrderByStripeSessionID(ctx context.Context, sessionID string) (domain.Order, error)
	UpdateOrderStatus(ctx context.Context, id uuid.UUID, status string) error

	// Shipping + requirements
	UpsertShippingDetails(ctx context.Context, orderID uuid.UUID, s domain.ShippingDetails) error
	UpsertPaymentRequirement(ctx context.Context, orderID uuid.UUID, r domain.PaymentRequirement) error

	// Payments
	AddPayment(ctx context.Context, orderID uuid.UUID, p domain.Payment) (domain.Payment, error)
	UpdatePaymentStatus(ctx context.Context, paymentID uuid.UUID, status string, paidAt *time.Time) error

	// Aggregates
	LoadOrderAggregate(ctx context.Context, id uuid.UUID) (domain.Order, error)
}

type Repository struct {
	db *store.Store
}

func New(pool *pgxpool.Pool) *Repository {
	return &Repository{
		db: store.New(pool),
	}
}

func (r *Repository) CreateOrder(ctx context.Context, order domain.Order, shipping domain.ShippingDetails, pricing domain.PaymentRequirement) {
	r.db.DoTx(ctx, func(ctx context.Context, q *generated.Queries) error {
		return nil
	})
}
