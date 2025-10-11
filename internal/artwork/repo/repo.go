package repo

import (
	"context"

	"github.com/google/uuid"
	"github.com/talmage89/art-backend/internal/artwork/domain"
	"github.com/talmage89/art-backend/internal/platform/db/store"
)

type Repo interface {
	ListArtworks(ctx context.Context) ([]domain.Artwork, error)
	GetArtworkDetail(ctx context.Context, id uuid.UUID) (*domain.Artwork, error)
	GetArtworkCheckoutData(ctx context.Context, ids []uuid.UUID) ([]domain.Artwork, error)
}

func New(db *store.Store) Repo {
	return &Postgres{db: db}
}
