package repo

import (
	"context"

	"github.com/art-vbst/art-backend/internal/auth/domain"
	"github.com/art-vbst/art-backend/internal/platform/db/generated"
	"github.com/art-vbst/art-backend/internal/platform/db/store"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

type Postgres struct {
	db *store.Store
}

func (p *Postgres) CreateUser(ctx context.Context, email string, passwordHash string) (*domain.UserWithHash, error) {
	var user *domain.UserWithHash

	err := p.db.DoTx(ctx, func(ctx context.Context, q *generated.Queries) error {
		params := generated.CreateUserParams{
			Email:        email,
			PasswordHash: passwordHash,
		}

		userRow, err := q.CreateUser(ctx, params)
		if err != nil {
			return err
		}

		user = p.toDomainUser(&userRow)
		return nil
	})

	if err != nil {
		return nil, err
	}

	return user, nil
}

func (p *Postgres) CreateRefreshToken(ctx context.Context, params *RefreshTokenCreateParams) (*domain.RefreshToken, error) {
	token := &domain.RefreshToken{}

	err := p.db.DoTx(ctx, func(ctx context.Context, q *generated.Queries) error {
		sessionID := uuid.New()
		if params.SessionID != nil {
			q.RevokeSessionRefreshTokens(ctx, *params.SessionID)
			sessionID = *params.SessionID
		}

		row, err := q.CreateRefreshToken(ctx, generated.CreateRefreshTokenParams{
			SessionID: sessionID,
			Jti:       params.Jti,
			UserID:    params.UserID,
			TokenHash: params.TokenHash,
			ExpiresAt: pgtype.Timestamp{Time: params.ExpiresAt, Valid: true},
		})

		token = p.toDomainRefreshToken(&row)
		return err
	})

	return token, err
}

func (p *Postgres) GetUser(ctx context.Context, id uuid.UUID) (*domain.UserWithHash, error) {
	user, err := p.db.Queries().GetUserByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return p.toDomainUser(&user), nil
}

func (p *Postgres) GetUserByEmail(ctx context.Context, email string) (*domain.UserWithHash, error) {
	user, err := p.db.Queries().GetUserByEmail(ctx, email)
	if err != nil {
		return nil, err
	}
	return p.toDomainUser(&user), nil
}

func (p *Postgres) GetRefreshTokenByJti(ctx context.Context, jti uuid.UUID) (*domain.RefreshToken, error) {
	row, err := p.db.Queries().GetRefreshTokenByJTI(ctx, jti)
	if err != nil {
		return nil, err
	}

	return p.toDomainRefreshToken(&row), nil
}

func (p *Postgres) RevokeToken(ctx context.Context, id uuid.UUID) error {
	return p.db.DoTx(ctx, func(ctx context.Context, q *generated.Queries) error {
		return q.RevokeRefreshToken(ctx, id)
	})
}

func (p *Postgres) RevokeUserTokens(ctx context.Context, userID uuid.UUID) error {
	return p.db.DoTx(ctx, func(ctx context.Context, q *generated.Queries) error {
		return q.RevokeAllUserRefreshTokens(ctx, userID)
	})
}

func (p *Postgres) toDomainUser(row *generated.User) *domain.UserWithHash {
	return &domain.UserWithHash{
		ID:           row.ID,
		Email:        row.Email,
		PasswordHash: row.PasswordHash,
	}
}

func (p *Postgres) toDomainRefreshToken(row *generated.RefreshToken) *domain.RefreshToken {
	return &domain.RefreshToken{
		Jti:       row.Jti,
		ID:        row.ID,
		UserID:    row.UserID,
		SessionID: row.SessionID,
		TokenHash: row.TokenHash,
		CreatedAt: row.CreatedAt.Time,
		ExpiresAt: row.ExpiresAt.Time,
		Revoked:   row.Revoked,
	}
}
