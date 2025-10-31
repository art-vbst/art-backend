package postgres

import (
	"context"
	"errors"

	"github.com/art-vbst/art-backend/internal/artwork/domain"
	"github.com/art-vbst/art-backend/internal/platform/db/generated"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

func (p *Postgres) UpdateArtworksForPendingOrder(ctx context.Context, orderId uuid.UUID, ids []uuid.UUID) error {
	return p.db.DoTx(ctx, func(ctx context.Context, q *generated.Queries) error {
		rows, err := q.UpdateArtworksForOrder(ctx, generated.UpdateArtworksForOrderParams{
			OrderID: pgtype.UUID{Bytes: orderId, Valid: true},
			Column2: ids,
		})
		if len(rows) != len(ids) {
			return errors.New("one or more artworks not found")
		}
		return err
	})
}

func (p *Postgres) UpdateArtworkStatuses(ctx context.Context, orderID uuid.UUID, status domain.ArtworkStatus) error {
	return p.db.DoTx(ctx, func(ctx context.Context, q *generated.Queries) error {
		params := generated.UpdateArtworkStatusParams{
			OrderID: pgtype.UUID{Bytes: orderID, Valid: true},
			Status:  status,
		}

		if _, err := q.UpdateArtworkStatus(ctx, params); err != nil {
			return err
		}

		return nil
	})
}
