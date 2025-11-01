package postgres

import (
	"context"

	"github.com/art-vbst/art-backend/internal/platform/db/generated"
	"github.com/google/uuid"
)

func (p *Postgres) DeleteArtwork(ctx context.Context, id uuid.UUID) error {
	return p.db.DoTx(ctx, func(ctx context.Context, q *generated.Queries) error {
		return q.DeleteArtwork(ctx, id)
	})
}
