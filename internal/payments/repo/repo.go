package repo

import (
	"context"

	"github.com/talmage89/art-backend/internal/payments/domain"
	"github.com/talmage89/art-backend/internal/platform/db/store"
)

type Repo interface {
	CreateOrder(ctx context.Context, order domain.Order) (domain.Order, error)
}

func New(db *store.Store) *Postgres {
	return &Postgres{db: db}
}
