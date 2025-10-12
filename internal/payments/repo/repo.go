package repo

import (
	"context"

	"github.com/google/uuid"
	"github.com/talmage89/art-backend/internal/payments/domain"
	"github.com/talmage89/art-backend/internal/platform/db/store"
)

type Repo interface {
	CreateOrder(ctx context.Context, order *domain.Order) (*domain.Order, error)
	UpdateOrderWithPayment(ctx context.Context, order *domain.Order, payment *domain.Payment) error
	DeleteOrder(ctx context.Context, orderID uuid.UUID) error
}

func New(db *store.Store) Repo {
	return &Postgres{db: db}
}
