package repo

import (
	"context"

	"github.com/art-vbst/art-backend/internal/artwork/domain"
	"github.com/art-vbst/art-backend/internal/artwork/repo/postgres"
	"github.com/art-vbst/art-backend/internal/platform/db/store"
	"github.com/google/uuid"
)

type Repo interface {
	ListArtworks(ctx context.Context) ([]domain.Artwork, error)
	CreateArtwork(ctx context.Context, body *domain.ArtworkPayload) (*domain.Artwork, error)
	GetArtworkDetail(ctx context.Context, id uuid.UUID) (*domain.Artwork, error)
	UpdateArtwork(ctx context.Context, id uuid.UUID, payload *domain.ArtworkPayload) (*domain.Artwork, error)
	DeleteArtwork(ctx context.Context, id uuid.UUID) error
	GetArtworkCheckoutData(ctx context.Context, ids []uuid.UUID) ([]domain.Artwork, error)
	UpdateArtworksForPendingOrder(ctx context.Context, orderID uuid.UUID, ids []uuid.UUID) error
	UpdateArtworkStatuses(ctx context.Context, orderID uuid.UUID, status domain.ArtworkStatus) error
}

func New(db *store.Store) Repo {
	return postgres.New(db)
}
