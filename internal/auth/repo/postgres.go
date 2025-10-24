package repo

import (
	"context"

	"github.com/art-vbst/art-backend/internal/auth/domain"
	"github.com/art-vbst/art-backend/internal/platform/db/generated"
	"github.com/art-vbst/art-backend/internal/platform/db/store"
)

type Postgres struct {
	db *store.Store
}

func (p *Postgres) CreateUser(ctx context.Context, email string, passwordHash string) (*domain.User, error) {
	var user *domain.User

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

func (p *Postgres) GetUserByEmail(ctx context.Context, email string) (*domain.User, error) {
	user, err := p.db.Queries().GetUserByEmail(ctx, email)
	if err != nil {
		return nil, err
	}

	return p.toDomainUser(&user), nil
}

func (p *Postgres) toDomainUser(row *generated.User) *domain.User {
	return &domain.User{
		ID:           row.ID,
		Email:        row.Email,
		PasswordHash: row.PasswordHash,
	}
}
