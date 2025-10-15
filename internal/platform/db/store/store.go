package store

import (
	"context"

	"github.com/art-vbst/art-backend/internal/platform/db/generated"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Store struct {
	pool *pgxpool.Pool
}

func New(pool *pgxpool.Pool) *Store { return &Store{pool: pool} }

func (s *Store) DoTx(ctx context.Context, fn func(ctx context.Context, q *generated.Queries) error) error {
	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}

	q := generated.New(tx)
	if err := fn(ctx, q); err != nil {
		_ = tx.Rollback(ctx)
		return err
	}

	return tx.Commit(ctx)
}

func (s *Store) Queries() *generated.Queries {
	return generated.New(s.pool)
}
