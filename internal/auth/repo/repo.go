package repo

import (
	"context"
	"time"

	"github.com/art-vbst/art-backend/internal/auth/domain"
	"github.com/art-vbst/art-backend/internal/platform/db/store"
	"github.com/google/uuid"
)

type Repo interface {
	CreateUser(ctx context.Context, email string, passwordHash string) (*domain.UserWithHash, error)
	GetUser(ctx context.Context, id uuid.UUID) (*domain.UserWithHash, error)
	GetUserByEmail(ctx context.Context, email string) (*domain.UserWithHash, error)
	CreateRefreshToken(ctx context.Context, params *RefreshTokenCreateParams) (*domain.RefreshToken, error)
	GetRefreshTokenByJti(ctx context.Context, jti uuid.UUID) (*domain.RefreshToken, error)
	RevokeToken(ctx context.Context, id uuid.UUID) error
	RevokeUserTokens(ctx context.Context, userID uuid.UUID) error
}

func New(db *store.Store) Repo {
	return &Postgres{db: db}
}

type RefreshTokenCreateParams struct {
	Jti       uuid.UUID
	UserID    uuid.UUID
	SessionID *uuid.UUID
	TokenHash string
	ExpiresAt time.Time
}
