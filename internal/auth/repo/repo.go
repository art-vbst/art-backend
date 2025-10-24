package repo

import (
	"context"

	"github.com/art-vbst/art-backend/internal/auth/domain"
	"github.com/art-vbst/art-backend/internal/platform/db/store"
)

type Repo interface {
	CreateUser(ctx context.Context, email string, passwordHash string) (*domain.User, error)
	// GetUserById(ctx context.Context, id uuid.UUID) (domain.User, error)
	GetUserByEmail(ctx context.Context, email string) (*domain.User, error)
	// CreateRefreshToken(ctx context.Context) (domain.RefreshToken, error)
	// GetRefreshTokenByHash(ctx context.Context, hash string) (domain.RefreshToken, error)
	// RevokeRefreshTokenByHash(ctx context.Context, has)
}

func New(db *store.Store) Repo {
	return &Postgres{db: db}
}
