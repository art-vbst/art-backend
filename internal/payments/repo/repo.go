package repo

import (
	"context"

	"github.com/art-vbst/art-backend/internal/payments/domain"
	"github.com/art-vbst/art-backend/internal/payments/repo/postgres"
	"github.com/art-vbst/art-backend/internal/platform/db/store"
	"github.com/google/uuid"
)

type Repo interface {
	CreateOrder(ctx context.Context, order *domain.Order) (*domain.Order, error)
	ListOrders(ctx context.Context) ([]domain.Order, error)
	GetOrder(ctx context.Context, id uuid.UUID) (*domain.Order, error)
	GetOrderPublic(ctx context.Context, id uuid.UUID, stripeSessionID *string) (*domain.OrderPublic, error)
	UpdateOrderStripeSessionID(ctx context.Context, id uuid.UUID, stripeSessionID *string) error
	UpdateOrderStatus(ctx context.Context, id uuid.UUID, status domain.OrderStatus) error
	UpdateOrderWithPayment(ctx context.Context, order *domain.Order, payment *domain.Payment) error
}

func New(db *store.Store) Repo {
	return postgres.New(db)
}
